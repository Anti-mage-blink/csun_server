# push.ps1 - 一键推送脚本
# 用法: .\push.ps1 "你的提交说明"
# 用法: .\push.ps1  （不传说明则自动生成带时间戳的说明）

param(
    [string]$Message = "auto commit: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')"
)

Set-Location $PSScriptRoot

Write-Host "==> 检查状态..." -ForegroundColor Cyan
git status -s

Write-Host "`n==> 添加文件..." -ForegroundColor Cyan
git add .

Write-Host "==> 提交: $Message" -ForegroundColor Cyan
git commit -m $Message

Write-Host "==> 推送到远程..." -ForegroundColor Cyan
git push

Write-Host "`n✅ 推送完成!" -ForegroundColor Green
