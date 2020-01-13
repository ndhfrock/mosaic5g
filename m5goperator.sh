#!/bin/bash

# prepare ENVs
export KUBECONFIG=/home/nadhif/.kube/config
export OPERATOR_NAME=m5g-operator
export MYNAME=${USER}
export MYDNS="140.118.31.99"

run_local(){
    operator-sdk up local --namespace=default
}

run_container(){
    case ${1} in
        start)
            kubectl create -f deploy/service_account.yaml
            kubectl create -f deploy/role.yaml
            kubectl create -f deploy/role_binding.yaml
            kubectl create -f deploy/operator.yaml
        ;;
        stop)
            kubectl delete -f deploy/service_account.yaml
            kubectl delete -f deploy/role.yaml
            kubectl delete -f deploy/role_binding.yaml
            kubectl delete -f deploy/operator.yaml
        ;;
    esac
}

init(){
    echo "Applying crd..."
    kubectl create -f deploy/crds/mosaic5g_v1alpha1_mosaic5g_crd.yaml
    sleep 3
    sudo mkdir -p /mnt/esdata
    sudo chmod 777 /mnt/esdata
    echo "Done, now run [local] or [container start] to create your operator"
}

clean(){
    kubectl delete -f deploy/crds/mosaic5g_v1alpha1_mosaic5g_crd.yaml
}

watch_dep(){
    sudo watch -n1 kubectl get deployment  
}

watch_pods(){
    sudo watch -n1 kubectl get pods -A -o wide
}

push_image() {
    operator-sdk build ndhfrock/m5goperator:v0.0.1
    sed -i 's|REPLACE_IMAGE|ndhfrock/m5goperator:v0.0.1|g' deploy/operator.yaml
    docker push ndhfrock/m5goperator:v0.0.1
}
go_inside() {
    kubectl exec -it $(kubectl get pods -l app=${1} -o custom-columns=:metadata.name) -- /bin/bash
}

main() {
    case ${1} in
        init)
            init
        ;;
        clean)
            clean
        ;;
        local)
            run_local
        ;;
        container)
            run_container ${2}
        ;;
        watch_deployment)
            watch_dep
        ;;	
        watch_pods)
            watch_pods
        ;;
	inside_pods)
	   go_inside ${2}
	;;
	push_image)
            push_image
        ;;
        *)
            echo "Bring up M5G-Operator for you"
            echo "[IMPORTANT] Please set up kubeconfig at the beginning of this script"
            echo ""
            echo "Usage:"
            echo "      m5goperator.sh init - Apply CRD to kubernetes cluster (Required for Operator)"
            echo "      m5goperator.sh clean - Remove CRD from cluster"
            echo "      m5goperator.sh local - Run Operator as a Golang app at local"
            echo "      m5goperator.sh container [start|stop] - Run Operator as a POD inside Kubernetes"
            echo "      m5goperator.sh watch_deployment - watch all running deployment, refreshed every 1 second"
            echo "      m5goperator.sh watch_pods - watch all running pods, refreshed every 1 second"
            echo "      m5goperator.sh push_image - push your operator image to dockerhub"
	    echo "      m5goperator.sh inside_pods [cn|ran|flexran] - access inside the pods"
            echo ""
            echo "Default operator image is ndhfrock/m5goperator"
        ;;
    esac

}
main "$@"
