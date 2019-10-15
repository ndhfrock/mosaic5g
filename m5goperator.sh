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
            kubectl apply -f deploy/service_account.yaml
            kubectl apply -f deploy/role.yaml
            kubectl apply -f deploy/role_binding.yaml
            kubectl apply -f deploy/operator.yaml
        ;;
        stop)
            kubectl delete -f deploy/service_account.yaml
            kubectl delete -f deploy/role.yaml
            kubectl delete -f deploy/role_binding.yaml
            kubectl delete -f deploy/operator.yaml
        ;;
    esac
}

deploy_operator_from_clean_machine(){
    echo "Start a fresh microk8s and deploy operator on it, tested with Ubuntu 18.04"
    echo "sudo without password is recommended"
    sudo snap install microk8s --classic --channel=1.14/stable
    sudo snap install kubectl --classic
    microk8s.start
    microk8s.enable dns
    # kubeconfig is used by operator
    sudo chown ${MYNAME} -R $HOME/.kube
    microk8s.kubectl config view --raw > $HOME/.kube/config
    # enable privileged
    sudo bash -c 'echo "--allow-privileged=true" >> /var/snap/microk8s/current/args/kubelet'
    sudo bash -c 'echo "--allow-privileged=true" >> /var/snap/microk8s/current/args/kube-apiserver'
    # Restart kube
    sudo systemctl restart snap.microk8s.daemon-kubelet.service
    sudo systemctl restart snap.microk8s.daemon-apiserver.service
    # Configure DNS if it's not working 
    # microk8s.kubectl -n kube-system edit configmap/coredns

}

init(){
    echo "Applying crd..."
    kubectl apply -f deploy/crds/mosaic5g_v1alpha1_mosaic5g_crd.yaml
    sleep 3
    echo "Done, now run [local] or [container start] to create your operator"
}

clean(){
    kubectl delete -f deploy/crds/mosaic5g_v1alpha1_mosaic5g_crd.yaml
}

break_down(){
    sudo snap remove microk8s 
    sudo snap remove kubectl 
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
        from_clean_machine)
            deploy_operator_from_clean_machine 
        ;;
        break_down)
            break_down
        ;;
        *)
            echo "Bring up M5G-Operator for you"
            echo "[IMPORTANT] Please set up kubeconfig at the beginning of this script"
            echo ""
            echo "Usage:"
            echo "      m5goperator.sh init - Apply CRD to k8s cluster (Required for Operator)"
            echo "      m5goperator.sh clean - Remove CRD from cluster"
            echo "      m5goperator.sh local - Run Operator as a Golang app at local"
            echo "      m5goperator.sh container [start|stop] - Run Operator as a POD inside Kubernetes"
            echo "      m5goperator.sh from_clean_machine - Install and run microk8s kubectl, then deploy operator on it (Tested with Ubuntu 18.04)"
            echo ""
            echo "Default operator image is tig4605246/m5g_operator:0.1"
        ;;
    esac

}
main "$@"
