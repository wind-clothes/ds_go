##MapReduce
MapReduce概述

* 上下文：关于多TB数据集的多小时计算

	* 例如分析爬网网页的图形结构

    * 仅适用于1000台电脑

    * 经常不是由分布式系统专家开发的

    * 分配可能非常痛苦，例如应对失败
* 总体目标：非专业程序员可以轻松拆分
    
	* 许多服务器的数据处理效率合理。
* 程序员定义了Map和Reduce功能
    
	* 顺序代码; 往往相当简单
* MR在具有巨大输入的1000台机器上运行功能
	
	* 并隐藏分布式细节
  
MapReduce的抽象视图

	输入分为M个文件
	输入1  - >Map - > a，1 b，1 c，1
    输入2  - >Map - > b，1
    输入3  - >Map - > a，1 c，1
                    | | |
                        |  - >减少 - > c，2
                        ----->减少 - > b，2

MR为每个输入文件调用Map（），生成一组k2，v2 “中间”数据,每个Map（）调用都是一个“任务”;MR收集给定k2的所有中间版本v2，并将其传递给Reduce()调用，最终输出是从Reduce（）设置的<k2，v3>存储在R输出文件中。

示例：wordcount

	输入是数千个文本文件
  	MAP（k，v）
    	把v分成单词
    	对于每个单词w
      	发射（w，“1”）
 	Reduce（k，v）
    发射（LEN（V））

MapReduce隐藏了许多痛苦的细节：

	在服务器上启动s / w
	跟踪哪些任务完成
	数据移动
  从故障中恢复

MapReduce性能增长比例：

* N台计算机可以获得Nx吞吐量。
    
	假设M和R是> = N（即大量的输入文件和输出键）。
    Maps（）可以并行运行，因为它们不进行交互。
    与Reduce（）相同。
* 所以你可以通过购买更多的电脑来获得更多的吞吐量
    
	而不是每个应用程序的特殊目的高效并行化。
    电脑比程序员便宜！

什么可能限制性能？
  中央处理器？记忆？磁盘？网络？


更多细节：

	Master：给Worker任务; 记录中间产出结果的地方
    M Map任务，R Reduce任务
    存储在GFS中的输入，每个Map输入文件的3个副本
    所有电脑都运行GFS和MR Worker
    比Worker更多的输入任务
    Master给每个worker一个Map任务
    Map Worker将中间密钥分为R个分区，在本地磁盘上
    减少通话，直到所有 Map 完成
    Master告诉Reducers从Map工作人员获取中间数据分区
    减少工作人员将最终输出写入GFS（每个Reduce任务一个文件）

详细设计如何减少网络慢的影响？

    Map输入从本地磁盘上的GFS副本读取，而不是通过网络读取。
      中间数据仅通过网络一次。
      Map工作人员写入本地磁盘，而不是GFS。
    中间数据被分割成包含许多键的文件。
      大网络传输效率更高。

他们如何获得良好的负载平衡？

    关键的缩放 -  N-1服务器等待1完成不好。
      但有些任务可能需要比其他任务更长的时间
    解决方案：比worker更多的任务。
      Master对完成以前任务的Worker提出新的任务。
      所以服务器的工作速度要比较慢的服务器要慢一点，在同一时间内完成。

容错？

比如：如果服务器在执行MR工作时崩溃怎么办？隐藏这个错误非常困难，为什么不重新执行这个工作呢？
    
    MR重新执行失败的Map函数和Reduce函数,他们是纯函数——他们不会改变数据输入、不会保持状态、不共享内存、不存在map和map，或者reduce和reduce之间的联系，

    所以重新执行也会产生相同的输出。纯函数的这个需求是MR相对于其他并行编程方案的主要限制，然后也是因为这个需求使得MR非常简单。

工作人员崩溃恢复的细节：

  * Map Worker 崩溃：

    Master看到Worker不再响应于ping
    崩溃的Worker的中间 Map 输出丢失
      但是每个 Reduce 任务都可能需要它！
    Master 重新运行，将任务扩展到其他GFS副本的输入。
    一些 Reduce 可能已经读过失败的 Worker 的中间数据。
      这里我们依赖于功能和确定性的Map（）！
    如果Reduces已经获取了所有中间数据，则master不需要重新运行Map
  * Reduce Worker崩溃

    完成的任务可以 - 存储在GFS中，具有副本。
    Master重新开始 Worker 对其他 Worker 的未完成任务。
  * 在编写输出时减少工作人员的崩溃。

    GFS具有原子重命名，防止在完成之前输出可见。
    所以主人可以安全地重新运行其他地方的Reduce任务。

其他故障/问题：

  * 如果主人给两个工作人员同一张Map（）任务怎么办？

    也许主人错误地认为一个工人死了。
    它会告诉减少工人只有其中一个。
  * 如果主人给两个工人同样的Reduce（）任务呢？

    他们都会尝试在GFS上编写相同的输出文件！
    原子GFS重命名防止混合; 一个完整的文件将可见。
  * 如果 Worker 很慢 - “掉队”怎么办？

    也许是由于flakey硬件。
    主人开始了最后几个任务的第二个副本。
  * 如果工作人员因h / w或s / w损坏而计算出错误的输出怎么办？

    太糟糕了！MR假设“故障停止”CPU和软件。
  * 如果主机崩溃怎么办？

什么应用程序不适合在MapReduce运行良好？
  不死所有的逻辑功能都时候 Map/shuffle/Reduce 模式。
  小数据，因为开销很高。例如不是网站的后端。
  对大数据的小更新，例如将一些文档添加到一个大的索引
  不可预测的读取（Map或Reduce都可以选择输入）
  更灵活的系统允许这些，但更复杂的模型。
