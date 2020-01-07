#!/bin/bash

APISERVER=`kubectl config view --minify -o jsonpath='{.clusters[0].cluster.server}'`
TOKEN=`kubectl get secret $(kubectl get serviceaccount default -o jsonpath='{.secrets[0].name}') -o jsonpath='{.data.token}' | base64 --decode `
   

apply_cr(){
   curl \
      -H "content-Type: application/json" \
      -H "Authorization: Bearer ${TOKEN}"\
      --insecure \
      -X POST ${APISERVER}/apis/mosaic5g.com/v1alpha1/namespaces/default/mosaic5gs \
      -d '{"apiVersion"		:	"mosaic5g.com/v1alpha1",
	   "kind"		:	"Mosaic5g",
	   "metadata"	:	{"name"				   :	"mosaic5g"},
	   "spec"		:	{
                "size"				         :	1,
					 "cnImage"			         :	"ndhfrock/oaicn:1.0",
					 "ranImage"			         :	"ndhfrock/oairan:1.1",
					 "flexRANImage"			   :	"mosaic5gecosys/flexran:0.1",
					 "mcc"				         :	"208",
					 "mnc"				         :	"93",
					 "eutraBand"			      :	"7",
					 "downlinkFrequency"		   :	"2685000000L",
					 "uplinkFrequencyOffset"	:	"-120000000",
					 "configurationPathofCN"	:	"/var/snap/oai-cn/current/",
					 "configurationPathofRAN"	:	"/var/snap/oai-ran/current/",
					 "snapBinaryPath"		      :	"/snap/bin/",
					 "hssDomainName"		      :	"cn",
					 "mmeDomainName"		      :	"cn",
					 "spgwDomainName"		      :	"cn",
					 "mysqlDomainName"	      :	"mysql",
					 "dns"				         :	"8.8.8.8",
					 "flexRAN"			         :	true,
					 "elasticsearch"		      :	false, 
					 "kibana"			         :	false, 
					 "droneStore"			      :	false, 
					 "rrmkpiStore"			      :	false, 
					 "flexRANDomainName"		   :	"flexran"}}'
}

delete_cr(){
   curl \
      -H "content-Type: application/json" \
      -H "Authorization: Bearer ${TOKEN}"\
      --insecure \
      -X DELETE ${APISERVER}/apis/mosaic5g.com/v1alpha1/namespaces/default/mosaic5gs/mosaic5g
}

apply_cr_slicing(){
   curl \
      -H "content-Type: application/json" \
      -H "Authorization: Bearer ${TOKEN}"\
      --insecure \
      -X POST ${APISERVER}/apis/mosaic5g.com/v1alpha1/namespaces/default/mosaic5gs \
      -d '{"apiVersion"		:	"mosaic5g.com/v1alpha1",
	   "kind"		:	"Mosaic5g",
	   "metadata"	:	{"name"				   :	"mosaic5g"},
	   "spec"		:	{
                "size"				         :	1,
					 "cnImage"			         :	"ndhfrock/oaicn:1.0",
					 "ranImage"			         :	"ndhfrock/oairanslicing:1.0",
					 "flexRANImage"			   :	"mosaic5gecosys/flexran:0.1",
					 "mcc"				         :	"208",
					 "mnc"				         :	"93",
					 "eutraBand"			      :	"7",
					 "downlinkFrequency"		   :	"2685000000L",
					 "uplinkFrequencyOffset"	:	"-120000000",
					 "configurationPathofCN"	:	"/var/snap/oai-cn/current/",
					 "configurationPathofRAN"	:	"/LTE_Mac_scheduler_with_network_slicing/targets/PROJECTS/GENERIC-LTE-EPC/CONF/",
					 "snapBinaryPath"		      :	"/snap/bin/",
					 "hssDomainName"		      :	"cn",
					 "mmeDomainName"		      :	"cn",
					 "spgwDomainName"		      :	"cn",
					 "mysqlDomainName"		   :	"mysql",
					 "dns"				         :	"8.8.8.8",
					 "flexRAN"			         :	true,
					 "elasticsearch"		      :	false, 
					 "kibana"			         :	false, 
					 "droneStore"			      :	false, 
					 "rrmkpiStore"			      :	false, 
					 "flexRANDomainName"		   :	"flexran"}}'
}

patch_12(){
   curl \
      -H "content-Type: application/json-patch+json" \
      -H "Authorization: Bearer ${TOKEN}"\
      --insecure \
      -X PATCH ${APISERVER}/apis/mosaic5g.com/v1alpha1/namespaces/default/mosaic5gs/mosaic5g \
      -d '[{"op"		:	"replace",
	    "path"		:	"/spec/cnImage",
	    "value"		:	"mosaic5gecosys/oaicn:1.2"}
	  ,{"op"		:	"replace",
	    "path"		:	"/spec/ranImage",
	    "value"		:	"mosaic5gecosys/oairan:1.2"}]'
}

patch_11(){
   curl \
      -H "content-Type: application/json-patch+json" \
      -H "Authorization: Bearer ${TOKEN}"\
      --insecure \
      -X PATCH ${APISERVER}/apis/mosaic5g.com/v1alpha1/namespaces/default/mosaic5gs/mosaic5g \
      -d '[{"op"		:	"replace",
	    "path"		:	"/spec/cnImage",
	    "value"		:	"mosaic5gecosys/oaicn:1.1"}
	  ,{"op"		:	"replace",
	    "path"		:	"/spec/ranImage",
	    "value"		:	"mosaic5gecosys/oairan:1.1"}]'
}

init(){
   kubectl apply -f defaultRole.yaml
}

main(){
   echo "---api.sh start---"
   case ${1} in
      init)
         init
      ;;
      apply_cr)
         apply_cr
      ;;
      apply_cr_slicing)
         apply_cr_slicing
      ;;
      delete_cr)
         delete_cr
      ;;
      patch_12)
         patch_12
      ;;
      patch_11)
         patch_11
      ;;
      *)
         echo "Commands: init apply_cr delete_cr patch_11 patch_12"
         echo "Bring up Mosaic 5G Custom Resource"
         echo "[Important] Always use api.sh init first to use this"
         echo ""
         echo "Usage:"
         echo "      api.sh init - Apply defaultRole to kubernetes cluster"
         echo "      api.sh apply_cr - Add custom resource deployment (uses snap oai-ran)"
	 echo "      api.sh apply_cr_slicing - Add custom resource deployment (uses samuel's oai-ran)"
         echo "      api.sh delete_cr - Delete all Custom Resource Deployment"
         echo "      api.sh patch_11 - Change to OAICN and OAIRAN Docker image tag 1.1"
         echo "      api.sh patch_12 - Change to OAICN and OAI RAN Docker image tag 1.2"
      ;;
   esac
   
   echo "---api.sh end---"
}

main ${1}
