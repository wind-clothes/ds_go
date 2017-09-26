## Spark

弹性分布式数据集：内存集群计算的容错抽象
Zaharia，Chowdhury，Das，Dave，Ma，McCauley，Franklin，Shenker，Stoica
NSDI 2012

今天：更多的分布式计算
  最后一个案例研究：火花
  为什么我们读Spark？
    广泛用于数据中心计算
    流行的开源项目，热启动（Databricks）
    支持迭代应用比MapReduce更好
    有趣的容错故事
    ACM博士论文奖

MapReduce使程序员的生活变得容易
  它处理：
    节点之间的通信
    分发代码
    安排工作
    处理失败
  但限制编程模式
    某些应用程序与MapReduce不兼容
    
许多算法是迭代的
  页面排名是典型的例子
    根据指向它的文档数量计算每个文档的排名
    该等级用于确定搜索结果中文档的位置
  在每次迭代中，每个文档：
    向其邻居发送r / n的贡献
       其中r是其等级，n是其邻居数。
    更新等级为alpha / N +（1鈭 alpha）* Sum（c_i），
       其总和超过了收到的捐款
       N是文件的总数。
  大计算：
    遍布世界各地的网页
    即使有很多机器需要很长时间

MapReduce和迭代算法
  MapReduce会在算法的一个迭代中很好
    每张地图都是文件的一部分
    减少更新特定文档的排名
  但下一次迭代怎么办？
    将结果写入存储
    开始新的MapReduce作业进行下一次迭代
    昂贵
    但容错

挑战
  更好的迭代计算编程模型
  良好的容错故事

一个解决方案：使用DSM
  适用于迭代编程
    等级可以在共享内存中
    作品可以更新和阅读
  错误容错
    典型的计划：记忆状态检查点
      每隔一小时记录一个检查点
    有两种方式：
      在计算时将共享内存写入存储
      失败后重做上次检查点后的所有工作
  Spark更多MapReduce的味道
    限制编程模式，但比MapReduce更强大
    良好的容错计划

更好的解决方案：将数据保存在内存中
  Pregel，Dryad，Spark等
  在火花中
    数据存储在数据集（RDD）中
    “坚持”记忆中的RDD
    下一次迭代可以参考RDD
    
其他机会
  互动数据探索
    对持久化的RDD运行查询
  喜欢有类似SQL的东西
    RDD上的连接运算符

Spark中的核心思想：RDD
  RDD是不可变的 - 您可以更新它们
  RDD支持转换和操作
  转换：从现有RDD计算新的RDD
    map，reduceByKey，filter，join，..
    转换是懒惰的：不要立即计算结果
    只是计算的描述
  行动：当需要结果时
    计数结果，收集结果，获得具体值

使用示例
  lines = spark.textFile（“hdfs：// ...”）
  errors = lines.filter（_。startsWith（“ERROR”））//懒惰！
  errors.persist（）//没有工作
  errors.count（）//计算结果的动作
  //现在错误在内存中被实现了
  //跨许多节点分区
  // Spark，将尝试保留在RAM中（当RAM已满时会溢出到磁盘）

再利用RDD
  errors.filter（_。包含（ “MySQL的”））。COUNT（） 
  //这将很快，因为重复使用以前片段计算的结果
  // Spark将在保存错误分区的计算机上安排作业

RDD的另一个重用
  errors.filter（_。含有（ “HDFS”））。图（_。分裂（ '\ T'）（3））。收集（）

RDD谱系
  Spark在动作中创建一个谱系图
  图形描述了使用转换的计算
    行 - >过滤器w错误 - >错误 - >过滤器w。HDFS  - > map  - >定时字段
  Spark使用谱系来调度作业
    在同一分区上的转型形成舞台
      例如，加入是一个阶段边界
      需要重新整理数据
    一个工作运行一个阶段
      一个阶段的管道改造
    计划RDD分区的作业

血统和容错
  *高效*容错的好机会
    假设一台机器失败了
    想重新计算*它的状态
    血统告诉我们什么是重新计算
      按照谱系识别所需的所有分区
      重新计算他​​们
  最后一个例子，找出缺少线的分区
    从孩子追溯到父母的血统
    所有的依赖都是“窄”
      每个分区依赖于一个父分区
    需要阅读丢失的行分区
      重新计算转换

RDD实现
  分区列表
  （父RDD，宽/窄依赖）
    narrow：取决于一个父分区（例如map）
    宽：取决于几个父分区（例如，join）
  计算功能（例如地图，连接）
  分区方案（例如，逐个文件）
  计算放置提示
  
每个变换采用（一个或多个）RDD，并输出变换后的RDD。

问：为什么RDD在其分区上携带元数据？A：所以转换
  这取决于多个RDD知道他们是否需要洗牌（宽）
  依赖）或不（狭义）。允许用户控制地点和减少
  洗牌。

问：为什么要区分狭义和广泛的依赖关系？答：万一
  失败。狭义的依赖只取决于需要的几个分区
  重新计算。广泛依赖可能需要整个RDD

示例：PageRank（从纸）：
  //将图形加载为（URL，outlinks）对的RDD
  val links = spark.textFile（...）。map（...）。persist（）//（URL，outlinks）
  var ranks = //（URL，rank）对的RDD
  for（i < -  1 to ITERATIONS）{
    //构建（targetURL，float）对的RDD
    //与每个页面发送的贡献
    val contribs = links.join（rank）.flatMap {
      （url，（links，rank））=> links.map（dest =>（dest，rank / links.size））
    }
    //通过URL总结贡献并获得新的排名
    ranks = contribs.reduceByKey（（x，y）=> x + y）
     .mapValues（sum => a / N +（1-a）* sum）
  }

PageRank的血统
  见图3
  每次迭代创建两个新的RDD：
    等级0，等级1等
    承诺0，承诺1等
  长谱图！
    容错的风险。
    一个节点失败，重新计算
  解决方案：用户可以复制RDD
    程序员通过“可靠”标志来持久化（）
     例如，每N次迭代调用ranks.persist（RELIABLE）  
    在内存中复制RDD
    使用REPLICATE标志，将写入稳定存储（HDFS）
  对性能的影响
   如果用户频繁出现w / REPLICATE，快速恢复，但执行速度较慢
   如果不经常，快速执行但恢复缓慢

问：是否坚持转型或行动？A：没有 它不创建一个
 新的RDD，并没有造成实质。这是一个指示
 调度。

问：通过调用没有标志的持久性，是否保证在出现故障的情况下
  RDD不需要重新计算？A：不，没有复制，所以一个节点
  拿着一个分区可能会失败。复制（在RAM中或稳定
  存储）是必需的

目前仅通过呼叫进行手动检查点维护。问：为什么要实施
  检查点？（这是昂贵的）A：长的谱系可能导致大的恢复
  时间。或者当有很宽的依赖性时，单个故障可能需要很多
  分区重新计算。

问：Spark可以处理网络分区吗？A：无法通信的节点
  调度程序将出现死机。网络的一部分可以从中接触到
  调度器可以继续计算，只要它有足够的数据启动
  从（如果所有分区的所有副本都无法达到，
  集群无法进展）

当内存不足时会发生什么？
  LRU（最近最少使用）在分区上
     首先不坚持
     然后持续（但它们将在磁盘上可用，确保用户不能超量内存）
  用户可以通过“持久优先”控制驱逐顺序
  没有理由不丢弃未持久化的分区（如果已经被使用）

性能
  降级到“几乎”MapReduce行为
  在图7中，100个Hadoop节点的k-means需要76-80秒
  在图12中，k-means表示25个Spark节点（内存中不允许分区）
    需要68.8差异可能是因为MapReduce会使用复制
    存储在减少后，但默认情况下Spark只会溢出到本地磁盘，
    没有网络延迟和副本上的I / O负载。没有建筑理由为什么
    火花将比MR慢

讨论
  Spark目标为批次，迭代应用程序
  Spark可以表达其他型号
    MapReduce，Pregel
  不能纳入新的数据
    但是请看Streaming Spark
  Spark不利于构建键/值存储
    像MapReduce等
    RDD是不可变的
    
参考
  http://spark.apache.org/
  http://www.cs.princeton.edu/~chazelle/courses/BIB/pagerank.htm
