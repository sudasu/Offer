# docker

--ulimit nofile=10240:40960    //指限制的打开的软硬文件数

--network=host                 //直接容器中端口和本地端口一一对应起来

-p loacl port:container port   //指定本地端口和容器端口对应起来