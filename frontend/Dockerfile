# syntax=docker/dockerfile:1.7

# ---------- 阶段1：构建 ----------
# 构建期用 Node 编译前端静态产物
FROM node:20-alpine AS builder
WORKDIR /app

# 先拷贝依赖描述文件，利用 Docker 层缓存加速重复构建
COPY package.json package-lock.json ./
RUN --mount=type=cache,target=/root/.npm \
    npm ci --registry=https://registry.npmmirror.com

# 拷贝源码并执行构建
COPY . .
RUN npm run build

# ---------- 阶段2：运行 ----------
# 运行期仅保留 Nginx + 静态产物，镜像小、安全
FROM nginx:1.27-alpine

# 拷贝构建产物
COPY --from=builder /app/dist /usr/share/nginx/html
# 拷贝 Nginx 配置
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
