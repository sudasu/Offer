# anaconda

## 修改镜像源

```
conda config --add channels https://mirrors.tuna.tsinghua.edu.cn/anaconda/pkgs/free/
conda config --set show_channel_urls yes
```

## 包管理

```
conda install pk_name            //安装
conda remove pg_name             //移除
conda update pg_name             //更新
conda list                       //查看
```

## 环境管理

```
conda create -n ev_name python=version     //创建
conda env remove ev_name                   //删除
conda env list                             //查看
conda activate ev_name                     //启动环境
conda deactivate                           //退出环境
```