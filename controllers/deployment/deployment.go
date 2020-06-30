package deployment

import (
	"strconv"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	v1 "zookeeper-controller/api/v1"
	"zookeeper-controller/controllers/common"
)


func AddEnvString(name string, value string) apiv1.EnvVar {
	var env apiv1.EnvVar
	env.Name = name
	env.Value = value
	return env
}

func AddPodEnv(name string, metaName string) apiv1.EnvVar {
	var env apiv1.EnvVar
	fieldref := apiv1.ObjectFieldSelector{}
	fieldref.APIVersion = "v1"
	fieldref.FieldPath = "metadata." + metaName
	field := apiv1.EnvVarSource{FieldRef: &fieldref}
	env.Name = name
	env.ValueFrom = &field
	return env
}

func AddHostVolume(path string, name string) apiv1.Volume {
	var (
		volume               apiv1.Volume
		volumeSource         apiv1.VolumeSource
		hostPathVolumeSource apiv1.HostPathVolumeSource
		hostType             apiv1.HostPathType
	)
	hostPathVolumeSource.Path = path
	hostType = apiv1.HostPathDirectoryOrCreate
	hostPathVolumeSource.Type = &hostType
	volumeSource.HostPath = &hostPathVolumeSource
	volume.Name = name
	volume.VolumeSource = volumeSource
	return volume
}

func CheckCPU(limitCPU string) string {
	if limitCPU == "" {
		limitCPU = "1000m"
	}
	return limitCPU
}

func CheckMemory(limitMemory string) string  {
	if limitMemory == "" {
		limitMemory = "2G"
	}
	return limitMemory
}

func DeployMentLogic(meta v1.ZookeeperSpec , req ctrl.Request) []appsv1.Deployment {

	var (
		deployList  []appsv1.Deployment

	)
	for _,v := range meta.NodeList {
		deployList = append(deployList,GetDeployment(meta,v,req))
	}

	return deployList
}

func GetDeployment(meta v1.ZookeeperSpec ,v v1.ZookeeperMeta, req ctrl.Request) appsv1.Deployment {
	var (
		//测试环境所有公共服务副本数固定是1
		replicas int32 = 1
		//RunAsUser int64 = 500
		env []apiv1.EnvVar
		volume []apiv1.Volume
		volumeMount []apiv1.VolumeMount
		containerPort []apiv1.ContainerPort
		someBool bool = true
	)
	// 判断值是否存在
	if meta.LogDir == "" {
		meta.LogDir = "/logs"
	}
	if meta.DataDir == "" {
		meta.DataDir = "/data"
	}
	if meta.NodeLogDir == "" {
		meta.NodeLogDir = "/opt/logs"
	}
	if meta.NodeDataDir == "" {
		meta.NodeDataDir = "/opt/data"
	}
	// 自定义lable标签
	lables := common.GetLables(v)
	deployment := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      v.Name,
			Namespace: req.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: lables,
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: lables,
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  v.Name,
							Image: meta.Image,
						},
					},
				},
			},
		},
	}
	// 添加容器端口
	containerPort = append(containerPort,common.AddContainerPort(v.Name,v.Port,2181))
	containerPort = append(containerPort,common.AddContainerPort("flower-port",v.FlowerPort,2888))
	containerPort = append(containerPort,common.AddContainerPort("leader-port",v.LeaderPort,3888))
	deployment.Spec.Template.Spec.Containers[0].Ports =containerPort
	// 添加容器环境变量
	env = append(env, AddEnvString("LANG", "en_US.UTF-8"))
	env = append(env, AddPodEnv("POD_NAME", "name"))
	env = append(env, AddPodEnv("POD_NAMESPACE", "namespace"))
	env = append(env, AddEnvString("Myid", strconv.Itoa(v.Myid)))
	env = append(env, AddEnvString("Port", strconv.Itoa(v.Port)))
	env = append(env, AddEnvString("FlowerPort", strconv.Itoa(v.FlowerPort)))
	env = append(env, AddEnvString("LeaderPort", strconv.Itoa(v.LeaderPort)))
	env = append(env, AddEnvString("NodeCount", strconv.Itoa(len(meta.NodeList))))
	for index,nodev := range meta.NodeList {
		index += 1
		env = append(env, AddEnvString("node"+strconv.Itoa(index), nodev.Name))
	}
	deployment.Spec.Template.Spec.Containers[0].Env = env
	// 处理容器日志持久化到node节点，默认日志路径 /logs node节点存放日志 ${meta.NodeLogDir} / namespace /deployment/ podname
	volume = append(volume, AddHostVolume(meta.NodeLogDir, "logdir"))
	volume = append(volume, AddHostVolume(meta.NodeDataDir, "datadir"))
	volumeMount = append(volumeMount,apiv1.VolumeMount{
		Name:             "logdir",
		//ReadOnly:         false,
		MountPath:        meta.LogDir,
		//SubPath:          "",
		//MountPropagation: nil,
		SubPathExpr:      "$(POD_NAMESPACE)/" + v.Name + "/$(POD_NAME)",
	})
	volumeMount = append(volumeMount, apiv1.VolumeMount{
		Name:             "datadir",
		MountPath:        meta.DataDir,
		//SubPath:          "",
		//MountPropagation: nil,
		SubPathExpr:      "$(POD_NAMESPACE)/" + v.Name,
	})
	deployment.Spec.Template.Spec.Volumes = volume
	deployment.Spec.Template.Spec.Containers[0].VolumeMounts = volumeMount
	// 添加健康监测
	//deployment.Spec.Template.Spec.Containers[0].LivenessProbe = &apiv1.Probe{
	//	Handler:             apiv1.Handler{
	//		TCPSocket: &apiv1.TCPSocketAction{
	//			Port: intstr.FromInt(2888),
	//		},
	//	},
	//	InitialDelaySeconds: 0, // 容器启动多长时间开始使用livebess探针
	//	TimeoutSeconds:      3,
	//	PeriodSeconds:       10,
	//	SuccessThreshold:    1,
	//	FailureThreshold:    3,
	//}
	//deployment.Spec.Template.Spec.Containers[0].ReadinessProbe = &apiv1.Probe{
	//	Handler:             apiv1.Handler{
	//		TCPSocket: &apiv1.TCPSocketAction{
	//			Port: intstr.FromInt(2888),
	//		},
	//	},
	//	//InitialDelaySeconds: 0,
	//	//TimeoutSeconds:      0,
	//	//PeriodSeconds:       0,
	//	//SuccessThreshold:    0,
	//	//FailureThreshold:    0,
	//}
	// 安全上下文
	deployment.Spec.Template.Spec.Containers[0].SecurityContext = &apiv1.SecurityContext{
		//Capabilities:             nil,
		//Privileged:               nil,  // 是否已root运行
		//SELinuxOptions:           nil,
		//WindowsOptions:           nil,
		//RunAsUser:                &RunAsUser,
		//RunAsGroup:               &RunAsUser,
		//RunAsNonRoot:             &someBool,
		ReadOnlyRootFilesystem:   &someBool,
		//AllowPrivilegeEscalation: nil,
		//ProcMount:                nil,
	}
	// 资源限制
	deployment.Spec.Template.Spec.Containers[0].Resources.Limits = apiv1.ResourceList{
		apiv1.ResourceMemory: resource.MustParse(CheckMemory(meta.Memory)),
		apiv1.ResourceCPU: resource.MustParse(CheckCPU(meta.CPU)),
	}
	deployment.Spec.Template.Spec.Containers[0].Resources.Requests = apiv1.ResourceList{
		apiv1.ResourceCPU: resource.MustParse("0"),
		apiv1.ResourceMemory: resource.MustParse("0"),
	}
	// 生命周期钩子预处理
	//deployment.Spec.Template.Spec.Containers[0].Lifecycle =


	return deployment
}