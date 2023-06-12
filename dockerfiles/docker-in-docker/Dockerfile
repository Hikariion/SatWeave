# 基于 Ubuntu 20.04 镜像
FROM ubuntu:20.04

# 安装依赖和更新 APT 缓存
RUN apt-get update && \
    apt-get install -y \
    apt-transport-https \
    ca-certificates \
    curl \
    gnupg-agent \
    software-properties-common

# 添加 Docker 的官方 GPG 密钥，并安装 Docker
RUN curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add - && \
    add-apt-repository \
   "deb [arch=amd64] https://download.docker.com/linux/ubuntu \
   $(lsb_release -cs) \
   stable" && \
   apt-get update && \
   apt-get install -y docker-ce docker-ce-cli containerd.io