快盘单向同步
===============

这是一个单向的同步工具，使用 **GO** 语言编写，可以检测目录的变化(写文件和删除文件),并将变化的文件同步到挂载了快盘的目录中, 如何在 **linux** 中挂载快盘请参考 https://github.com/alex8224/fuse_for_kuaipan 项目


依赖
~~~~~

inotify


用法
~~~~~~

./kuaipan_syncer_x64 被监控的目录 快盘挂载的目录

