package service

import (
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	v1 "zookeeper-controller/api/v1"
	"zookeeper-controller/controllers/common"
)


func ServiceLogic(meta v1.ZookeeperSpec , req ctrl.Request) []apiv1.Service {
	var (
		serviceList []apiv1.Service
	)
	for _,v := range meta.NodeList {
		serviceList = append(serviceList,GetService(meta,v,req))
	}
	return serviceList
}


func GetService (meta v1.ZookeeperSpec,v v1.ZookeeperMeta ,req ctrl.Request) apiv1.Service{
	var (
		servicePort []apiv1.ServicePort
	)
	servicePort = append(servicePort,common.AddServicePort(v.Name,v.Port,2181,v.Port))
	servicePort = append(servicePort,common.AddServicePort("flower-port",v.FlowerPort,2888,v.FlowerPort))
	servicePort = append(servicePort,common.AddServicePort("leader-port",v.LeaderPort,3888,v.LeaderPort))
	if v.ServiceType == "" {
		v.ServiceType = "ClusterIP"
	}
	service := apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: v.Name,
			Namespace: req.Namespace,
		},
		Spec:       apiv1.ServiceSpec{
			Selector: common.GetLables(v),
			Type: v.ServiceType,
			ClusterIP: "None",
			Ports: servicePort,
			SessionAffinity: "ClientIP",
		},
		Status:     apiv1.ServiceStatus{},
	}


	return service
}
