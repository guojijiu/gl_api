@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

REM Cloud Platform API Windows部署脚本

echo.
echo ========================================
echo    Cloud Platform API 部署脚本
echo ========================================
echo.

REM 检查Go是否安装
go version >nul 2>&1
if errorlevel 1 (
    echo [错误] Go未安装，请先安装Go 1.21+
    pause
    exit /b 1
)

REM 检查Docker是否安装
docker version >nul 2>&1
if errorlevel 1 (
    echo [警告] Docker未安装，将使用本地部署模式
    set USE_DOCKER=false
) else (
    echo [信息] Docker已安装
    set USE_DOCKER=true
)

REM 检查Docker Compose是否安装
docker-compose version >nul 2>&1
if errorlevel 1 (
    if "!USE_DOCKER!"=="true" (
        echo [警告] Docker Compose未安装，将使用Docker单容器模式
    )
)

REM 解析参数
if "%1"=="" (
    set DEPLOY_MODE=local
) else (
    set DEPLOY_MODE=%1
)

REM 根据部署模式执行相应操作
if "%DEPLOY_MODE%"=="local" (
    call :deploy_local
) else if "%DEPLOY_MODE%"=="docker" (
    if "!USE_DOCKER!"=="true" (
        call :deploy_docker
    ) else (
        echo [错误] Docker未安装，无法使用Docker部署
        pause
        exit /b 1
    )
) else if "%DEPLOY_MODE%"=="production" (
    call :deploy_production
) else (
    echo [错误] 未知的部署模式: %DEPLOY_MODE%
    echo 支持的部署模式: local, docker, production
    pause
    exit /b 1
)

echo.
echo 部署完成！
pause
exit /b 0

:deploy_local
echo [信息] 开始本地部署...
echo [信息] 下载Go依赖...
go mod download
go mod tidy

echo [信息] 构建应用...
go build -o bin\app.exe .

if not exist .env (
    echo [信息] 创建环境配置文件...
    copy env.example .env
    echo [警告] 请编辑 .env 文件配置数据库等信息
)

echo [信息] 本地部署完成！
echo [信息] 运行命令: bin\app.exe
goto :eof

:deploy_docker
echo [信息] 开始Docker部署...
echo [信息] 构建Docker镜像...
docker build -t cloud-platform-api .

REM 检查是否有docker-compose
docker-compose version >nul 2>&1
if not errorlevel 1 (
    echo [信息] 使用Docker Compose启动服务...
    docker-compose up -d
    
    echo [信息] Docker部署完成！
    echo [信息] 访问地址: http://localhost:8080
    echo [信息] 查看日志: docker-compose logs -f app
) else (
    echo [信息] 使用Docker单容器模式...
    docker run -d --name cloud-platform-api -p 8080:8080 -e SERVER_PORT=8080 -e SERVER_MODE=production -e DB_DRIVER=sqlite -e DB_DATABASE=cloud_platform.db cloud-platform-api
    
    echo [信息] Docker部署完成！
    echo [信息] 访问地址: http://localhost:8080
    echo [信息] 查看日志: docker logs -f cloud-platform-api
)
goto :eof

:deploy_production
echo [信息] 开始生产环境部署...

REM 检查环境变量
if "%PRODUCTION_DB_HOST%"=="" (
    echo [错误] 生产环境需要设置数据库环境变量
    echo [错误] 请设置: PRODUCTION_DB_HOST, PRODUCTION_DB_PASSWORD 等
    pause
    exit /b 1
)

echo [信息] 构建生产版本...
set CGO_ENABLED=0
set GOOS=linux
go build -a -installsuffix cgo -o bin\app .

echo [信息] 生产环境部署完成！
echo [信息] 运行命令: bin\app
goto :eof

