# user-rpc 服务配置
SERVER_NAME := user
SERVER_TYPE := rpc
DOCKERFILE := deploy/dockerfile/Dockerfile_$(SERVER_NAME)_$(SERVER_TYPE)_dev
IMAGE_NAME := $(SERVER_NAME)-$(SERVER_TYPE)
IMAGE_TAG := dev

# 构建开发环境镜像
release-test:
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) -f $(DOCKERFILE) .

# 本地运行容器
run: release-test
	docker run --rm -p 8080:8080 $(IMAGE_NAME):$(IMAGE_TAG)

# 清理镜像
clean:
	docker rmi $(IMAGE_NAME):$(IMAGE_TAG) 2>/dev/null || true

.PHONY: release-test run clean
