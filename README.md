# feishu-rss-release-robot

## 背景
继飞书 RSS 应用因不可抗力下架后，一时没寻找到合适的 IFTTT 类的工具可以直接打通订阅信息发到飞书群的。  
一些比较重要的仓库发版期望直接通知到群里所有人，拉齐共识，花了一个晚上写了个高度定制化的 robot 服务，可能只适合订阅 GitHub Repository 的 release 信息 🤣

## 使用
将 `src/config.json` 涉及到的配置替换成申请的相关和订阅的仓库 release.atom 链接，推荐走 gitee 镜像更快。  


![demo](https://img.lastwhisper.cn/feishu-rss-release-robot-demo.png)
