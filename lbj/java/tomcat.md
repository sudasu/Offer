# [Tomcat](https://tomcat.apache.org/tomcat-8.5-doc/index.html)

## 概述

Tomcat 10以及以后的版本遵循Jakarta EE规范开发，而Tomcat 9及之前的版本遵循Java EE规范开发。

## 目录管理

目录结构：

```
/bin             //运行脚本和启动程序
/conf            //配置文件
/lib             //核心库文件
/logs            //日志目录
/temp            //用于jvm临时文件使用的临时目录
/webapps         //用户自定义装载的程序
/work            //部署web程序使用的临时工作目录
```

使用环境变量控制多个tomcat实例：

```
CATALINA_HOME:配置多实例的公用文件目录，一般为bin,lib,可以通过升级公用文件目录的内容实现tomcat升级。其中bin目录里面的tomcat-juli.jar建议不要修改，如果想自己替换日志系统，建议替换BASE里的jar。
CATALINA_BASE:配置每个实例的私有文件目录，一般为conf,work,logs,temp,webapps等,可以启动多个独立实例。以上目录都是默认有的，没有会自动创建但是conf目录必须有server.xml和web.xml配置文件，否则会启动失败。如果bin和lib都有的话，先读base再读home(如果只有一个实例则HOME，BASE一致)
```

运行时动态修改目录地址:

```
Unix: CATALINA_BASE=/tmp/tomcat_base1 catalina.sh start
Windows: CATALINA_BASE=C:\tomcat_base1 catalina.bat start
```

## Ant匹配

```
?    匹配任何单字符
*    匹配0或者任意数量的字符
**   匹配0或者更多的目录
```

## 类加载器

tomcat提供多种类加载器，以实现容器的不同部分和多个不同web应用运行在一个容器里面。

### Bootstrap

Java虚拟机自带的类加载器，用于加载基础运行时需要的类，和基于System Extensions目录(/jre/lib/ext)的类。当然，JVMs根据其实现可能不止一个类加载器如extension classloader，作为类加载器也不一定可见。

### System



## Connector

### HTTP Connector

HTTP Connector考虑的是更好的整体性能与更低的延时。对于集群而言，需要使用Http负载均衡器来指导流量的分发(Apache HTTP Server 2.x默认被包含)。tomcat支持mod\_proxy作为负载均衡器，但是基于AJP协议的集群性能明显更好。

### AJP Connector

Tomcat最主要的功能是提供Servlet/JSP容器，尽管它也可以作为独立的Java Web服务器，它在对静态资源（如HTML文件或图像文件）的处理速度，以及提供的Web服务器管理功能方面都不如其他专业的HTTP服务器，如IIS和Apache服务器。AJP协议基于tcp协议通过二进制格式来传输可读性文本，tomcat通过使用AJP连接用来跟其他HTTP服务器建立连接，保证仅处理Servlet/JSP相关的内容。
