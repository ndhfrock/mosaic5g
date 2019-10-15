#!/bin/bash

KUBE_VERSION="1.15.1-00"

install_req(){
    sudo apt install -qy kubeadm=${KUBE_VERSION} kubelet=${KUBE_VERSION} kubectl=${KUBE_VERSION}
    sudo apt-mark hold kubeadm kubelet kubectl
    sudo swapoff -a
}

remove_req(){
    sudo apt-mark unhold kubeadm kubelet kubectl
    sudo apt remove -qy kubeadm kubelet kubectl
}

start(){
    # Create Single Node Kubernetes automatically, run this with normal user
    STARTUP_TYPE=${1}
    echo ${STARTUP_TYPE}
    sudo rm -r .kube/
    sudo ln -s /run/resolvconf/ /run/systemd/resolve
    sudo swapoff -a

    echo "Creating Kubernetes with CNI: ${STARTUP_TYPE}"
    # POD network is different between each cni plugin
    if [ ${STARTUP_TYPE} == "flannel" ]; then
        sudo kubeadm init --pod-network-cidr=10.244.0.0/16
    elif [ ${STARTUP_TYPE} == "calico" ]; then
        sudo kubeadm init --pod-network-cidr=192.168.0.0/16
    fi

    echo "Adding config to ${HOME}"
    echo "Sleep wait 1 sec"
    sleep 1

    mkdir -p $HOME/.kube
    sudo cp /etc/kubernetes/admin.conf $HOME/.kube/config
    sudo chown 1000:1000 $HOME/.kube/config

    echo "Sleep to wait master booting up"
    sleep 1

    #flannel
    if [ ${STARTUP_TYPE} == "flannel" ]; then
        echo "Apply flannel"
        kubectl apply -f https://raw.githubusercontent.com/coreos/flannel/master/Documentation/kube-flannel.yml 
    elif [ ${STARTUP_TYPE} == "calico" ]; then 
        echo "Apply calico"
        kubectl apply -f \
        https://docs.projectcalico.org/v3.6/getting-started/kubernetes/installation/hosted/kubernetes-datastore/calico-networking/1.7/calico.yaml
    fi

    echo "wait for kubernetes is ready"
    sleep 10

    echo "Schedule POD on Master"
    kubectl taint node ${HOSTNAME} node-role.kubernetes.io/master:NoSchedule-

    echo "Done, good to go"
}


stop(){
    echo "Stopping Kubernetes"
    sudo kubeadm reset -f 
    echo "Kubernetes Stopped by stop()"
}



main() {
    if [ `id -u` = "0" ]; then
        echo "please run this as normal user and  set up sudo without password"
        return -1
    fi
    case ${1} in
        start)
            start ${2}
            echo "Extras for Ubuntu 16.04 (If core-dns is not working):"
            echo "1、kubectl edit cm coredns -n kube-system"
            echo "2、delete ‘loop’ ,save and exit"
            echo "3、kubectl -n kube-system delete pod -l k8s-app=[kube-dns|core-dns]"
        ;;
        stop)
            stop
        ;;
        install_req)
            install_req
        ;;
        remove_req)
            remove_req
        ;;
        *)
            echo "Requirement:"
            echo "  1. Set up sudo without password and run this script as normal user"
            echo "  2. Kubernetes source repo must be added. Check here: https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/install-kubeadm/#installing-kubeadm-kubelet-and-kubectl"
            echo "----"
            echo "Description:"
            echo "  This script uses kubeadm to create a custom Kubernetes with calico/flannel CNI plugin"
            echo "  tested with 18.04/16.04 Ubuntu; Using Kubernetes version ${KUBE_VERSION}"
            echo "Usage:"
            echo "  createk8s.sh install_req ---- Install kubeadm, kubelet and kubectl. Disable swap"
            echo "  createk8s.sh remove_req  ---- remove kubeamd, kubelet and kubectl"
            echo "  createk8s.sh start [calico|flannel] ---- Create a k8s master with CNI installed"
            echo "  createk8s.sh stop ---- break down k8s master"
            echo "  createk8s.sh start simple ---- For minimal setup (no cni plugin)"
            echo "Note: You still need to add other worker nodes manually. "
        
    esac

}
main ${1} ${2}
