#!/bin/bash
reso_addr='registry.cn-hangzhou.aliyuncs.com/easy-chat/social-rpc-dev'
tag='latest'


container_name="easy-chat-social-rpc-test"

docker stop ${container_name}

docker rm ${container_name}

docker rmi ${reso_addr}:${tag}

docker pull ${reso_addr}:${tag}


# 如果需要指定配置文件的
docker run --network host --name=${container_name} -d ${reso_addr}:${tag}