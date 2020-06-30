package common

import (
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	v1 "zookeeper-controller/api/v1"
)

func GetLables(meta  v1.ZookeeperMeta) map[string]string {
	lables := map[string]string{}
	lables["serviceName"] = meta.Name
	return lables
}

func AddContainerPort(portName string,listen_port int,default_port int) apiv1.ContainerPort{
	var (
		containerPort apiv1.ContainerPort
	)
	containerPort.Name = portName
	containerPort.Protocol = "TCP"
	if listen_port == 0 {
		containerPort.ContainerPort = int32(default_port)
	}else {
		containerPort.ContainerPort = int32(listen_port)
	}
	return containerPort
}

func AddServicePort(portName string,listen_port int,default_port int,targetPort int) apiv1.ServicePort {
	var (
		servicePort apiv1.ServicePort
	)
	servicePort.Name = portName
	servicePort.Protocol = "TCP"
	if listen_port == 0 {
		servicePort.Port = int32(default_port)
		servicePort.TargetPort = intstr.FromInt(targetPort)
	}else {
		servicePort.Port = int32(listen_port)
		servicePort.TargetPort = intstr.FromInt(targetPort)
	}
	return servicePort
}