# git指令

## 常用基础

新增分支：`git branch branchname [版本号]`  //默认为当前分支创建，可根据log指定版本创建分支  
基于历史版本切换新增：`git checkout -b newbranch basebranch`       //基于分支或者版本号创建
删除分支：`git branch -d/-D [branchname]` -d不能删除未合并完成等中间状态分支，需要使用-D  
切换和文件忽略。  
删除远端分支:`git push origin --delete branchname`  
提交指令：git commit -a     			//-a表示提交修改和删除。新增还是需要add  
		   git commit --amend		//表示补充提交，在上次提交的基础上补充，而不会生成新的提交
新增指令：git add [-u|-A|.]   			//-u表示update。提交修改和删除，.是修改和新建，-A是所有删除，替换，修改，新增  

推送指令：git push origin branchname(不存在的话，自动创建远端分支)  //-f会强行覆盖掉冲突  
查看映射关系：git branch -vv  
建立远端映射：git push --set-upstream origin branchname  
忽略:对于未tracked文件，直接在.ignore文件添加即可。但已被tracker的文件，需要使用rm --cached移除索引再提交。但如果是需要被track，又不想让自己本地的更新被察觉，可以使用git update-index --assume-unchanged -- <file>指令，注意只能是file,而不能是filepath，如想取消可以使用no-assume-unchanged。事实证明不好使，总是有隐患的，比如reset hard后容易丢失。  
合并与变基:git rebase					//一般为了保证本地分支的整洁，在合并最新develop分支时都采用变基的方式而不
是merge

```
git rebase upstream [selectedbranch]		 //如果selected为空则将upstream分支基于当前分支rebase，如果不是为空则upsteam以selectedbranch为基础rebase且切换至selectedbranch。rebase是将别人的修改添加在你的分支的后面，方便你查看自己的分支log。
git rebase --onto abranch bbranch cbranch		//你甚至可以将依赖于b的c截断出来拼接到a上。
git rebase --continue|skip|abore			//在rebase过程中如果遇见了conflicts，有三种选择方式，要么解决冲突后continue。要么直接覆盖掉skip，要么取消rebase，切忌不要手贱commit否则使用abort退回。
git rebase -i head^2						//例子，交互式将当前修改压缩至上两个版本
```

stash指令：

```
git stash          				//执行存储 可 + save "save message"
git stash list 				//列出stash缓存列表
git stash show				//显示stash做了哪些改动，默认show第一个，如果需要看其他的需加上@{$num}
git stash pop				//恢复缓存的工作目录，会从list中去除
git stash apply				//应用第一个缓存，不会删除，如要使用其他的需加上 stash@{$num}
git stash drop stash@{$num}	//删除缓存
git stash clear				//删除所有缓存
```
回退指令:关于reset head^^后reset <head> --soft回原头节点的思考。混合型倒退head^^版本后，文件带回去了索引没带回去，所以status会显示很多修改。而软重置会原头节点，将文件带回去的同时将两个版本前的索引带回去了，status仍然会显示很多修改。这时千万不能按照索引来做修改，比如clean掉未track的，modify文件之类这样操作后的原head就变成head^^的状态了。如果不小心如此操作了，可以选择硬重置复原文件和索引，而正确操作是此时只需reset head，继续将当前状态的文件带回head但不带索引即将索引重置成head版本，即可。(最好瞎操作前，先commit或者stash，不然依靠版本控制重置文件，那存储之前的文件可就都没了)

```
git checkout [basebranch]-- filepath			//将工作区中的修改回退至暂存区的索引状态(add之前有效)，
												//相比较之下，reset filepath可根据各版本的情况去修改filepath下的索引，文件状态。更新一波，
												//hard不允许在filepath下reset，只能mixed，可能soft可以。

git checkout -b [localbranch] origin/[remotebranch] //拉取远端分支，且自动建立映射

git restore filepath        					//使用起来等同于checkout,可以加--staged参数将暂存区的索引退回
git revert HEAD^					//反向操作指定的版本以达到撤销，然后该操作可以当作commit提交。相比较reset的直线撤回，revert可以反做中间版本以用来撤回。其中revert默认是带编辑信息的提交，如需要不提交自己提交，使用 -n。如果遇见合并分支则无法撤回需要使用-m 指定分支进行撤回。撤回操作仍会导致conflicts，需要手动解决。
git reset [--soft|mixed|hard] [版本号]	//默认为混合型，且head版本 soft会回到暂存区已更新只差commit状态
mixed回回到暂存区未更新，工作区已更新的状态，还需add，hard则会直接丢失所有更改，可以无视现在的未提交直接重置。

快捷回退版本号
git reset						//混合型回退当前版本
git reset --hard HEAD~3 		//回退3个版本,0表示当前版本
git reset HEAD^^				//回退2个版本，表示切换到从当前开始的第三个版本
```

打扫指令:相比较reset，clean从当前目录清除未被track的文件，reset清除被track的文件。一般硬重置后都可能没有需要clean的(目前发现reset中途包含有merge,则需要clean)，混合重置可能会有没有add的，可以使用clean清理。

```
git clean 各后缀参数说明： 
-n 					//显示将要被删除的文件
-d	[path]			//如果没有给path，为了避免删除过多文件，并不会递归删除。使用-d递归删，如加了path则没啥影响
-f 	 				//一般默认配置都是不允许直接删除，必须使用-f来保证删除或者-i交互式删除
-i					//提供交互式问答，以保证删除的准确性
-x					//删除所有为track的文件，包括.gitignore里面的 
-X					//删除只包含在.gitignore里面的
```

rm指令

```
git rm <path> 指令说明:其中rm将会非递归的直接移除git目录的索引和文件
-f					//如index与文件状态不一致，需要使用-f强制覆盖移除
-r					//递归移除目录文件
--cached			//仅移除目标的index，即可用来ignore tracked文件
-n					//查看将要移除的文件，但不执行
```
log指令

```
git log 				//查看提交历史，以便确定回退到哪个版本
git log --oneline --graph --all	//查看分支图
git log -g			//会以标准格式输出reflog
git reflog 			//查看命令历史，以便回到hard reset后的版本
```

gc指令：工作时遇见了unable to update local ref的问题，使用gc解决了(有时间了再研究)
```
```

remote指令：维护远端信息

```
git remote -v                      //查看远端信息
git remote set-url origin [url]    //设置远端信息
git ls-remote --tag origin         //查看远端tag
```

tag指令： 标签，稳定上线版本.我们常常在代码封板时,使用git创建一个tag ,这样一个不可修改的历史代码版本就像被我们封存起来一样,不论是运维发布拉取,或者以后的代码版本管理,都是十分方便的

```
git tag                         //查看本地tag
git tag -d <tagName>            //删除tag
git tag <newTagName>		    //新建tag
git tag -a <tagName> <commitId> //以历史的某个提交id创建tag
git push origin <newTagName>    //推送到远程仓库
```