## FaRM QA
问：目前使用FaRM的系统有哪些？

A：只是FaRM。它不在生产中使用; 这是一个最近的研究原型。我怀疑它会影响未来的设计，也许本身就会发展成为一个生产系统。

问：为什么微软研究部发布这个？谷歌，Facebook，雅虎等研究部门也是如此，为什么总是揭示这些新系统的设计？技术进步肯定是更好的，但似乎（至少在短期内），保持这些设计是最有利的。

答：这些公司只发表一些关于他们撰写的软件的一小部分的论文。他们发表的一个原因是这些系统由具有学术背景（即谁拥有博士学位）的人部分开发，他们认为他们生活中的一部分任务是帮助世界了解他们发明的新想法。他们为自己的工作感到自豪，并希望人们欣赏它。另一个原因是这些论文可能会帮助公司吸引顶尖人才，因为这些论文表明，智力上有趣的工作正在发生。

问：在性能方面，FaRM与Thor相比如何？看起来容易理解

A：FaRM比Thor快多个数量级。

问：FaRM是否真的表明在分布式系统的一致性/可用性方面必须妥协的结束？

答：我不怀疑。例如，如果您愿意进行非事务性单面RDMA读写操作，那么可以像FaRM完成交易一样快速完成约5倍的操作。也许很少有人希望今天有5倍的表现，但有可能会有一天。另一方面，对于地理分布数据的交易（或一些良好的语义）来说，FaRM似乎并不重要，因此有很大的兴趣。

问：本文并没有把焦点放在FaRM的底线上。使用FaRM的一些最大的缺点是什么？

A：这是一些猜测。
数据必须适合RAM。
如果有许多冲突的交易，使用OCC并不是很大。
事务API（在其NSDI 2014文章中描述）看起来很难使用，因为回调在回调中返回。
应用程序代码必须紧密地交织执行的应用程序事务和轮询RDMA NIC队列和日志以获取来自其他计算机的消息。
应用程序代码在执行将最终中止的事务时可以看到不一致。例如，如果事务在提交事务覆盖对象的同时读取一个大对象。风险在于，如果没有防御性写入，应用程序可能会崩溃。
应用程序可能无法使用自己的线程，因为FaRM引脚线程到内核并使用所有内核。当然，FaRM是一个研究原型，只是为了探索一些新的想法。它不是用于一般用途的成品。如果人们继续这种工作，我们最终可能会看到FaRM的后裔，粗糙的边缘较少。

问：RDMA与RPC有何不同？什么是单面RDMA？

答：RDMA是在一些现代NIC（网络接口卡）中实现的一个特殊功能。NIC寻找通过网络到达的特殊命令数据包，并执行命令本身（并且不会给CPU发送数据包）。这些命令指定存储器操作，例如向地址写入值或从地址读取，并通过网络发送值。另外，RDMA NIC允许应用程序代码直接与NIC硬件通信，发送特殊的RDMA命令数据包，并在“硬件ACK”数据包到达时通知接收NIC已执行该命令。

“单面”是指一台计算机中的应用程序代码使用这些RDMA NIC直接在另一台计算机上读取或写入内存而不涉及其他计算机的CPU的情况。第4节/图4中的FaRM的“验证”阶段仅使用单面读取。FaRM有时使用RDMA作为实现类似RPC的方案来与接收计算机上运行的软件通信的快速方式。发送方使用RDMA将请求消息写入接收器的FaRM软件轮询（定期检查）的存储器区域; 接收方以相同的方式发送其回复。FaRM“锁定”阶段以这种方式使用RDMA。

传统RPC要求应用程序对本地内核进行系统调用，本地内核要求本地NIC发送数据包。在接收计算机上，NIC将数据包写入存储器中的队列，并破坏接收计算机的内核。内核将数据包复制到用户空间，并使应用程序看到它。接收应用程序相反发送回复（系统调用到内核，内核与NIC通信，另一方面是NIC中断其内核，＆c）。这一点是为每个RPC执行大量的代码，并不是很快。

RDMA的优点是速度快。单面RDMA读或写只需要微秒的1/18（图2），而传统的RPC可能需要10微秒。即使FaRM使用RDMA进行消息传递也比传统的RPC快得多，因为当涉及到双方的CPU时，双方都不涉及内核。

问：FaRM的表现好像主要来自硬件。除了哈尔滨配件，FaRM设计的重点是什么？

A：这是真的，FaRM的一个原因是硬件是快的。但是硬件已经存在了很多年了，但没有人想出如何将所有的部件放在一起，以真正利用硬件的潜力。FaRM做得很好的一个原因是他们同时投入了大量精力来优化网络，持久存储和使用CPU; 许多以前的系统已经优化了一个但不是全部。一个具体的设计点是FaRM对许多交互使用快速单面RDMA（而不是较慢的完整RPC）的方式。

问：该文件指出，FaRM利用了最近的硬件趋势，特别是UPS的成长使DRAM内存不稳定。有没有我们读过的这些系统是否被创造出来呢？例如，Raft的现代实施不再需要担心持续存在非易失性状态，因为它们可以配备UPS？

A：想法很旧 例如Harp复制文件服务在20世纪90年代初使用它。许多存储系统以其他方式使用电池（例如，在RAID控制器中），以避免不必等待磁盘写入。然而，FaRM使用的电池设置并不常见，所以必须通用的软件不能依赖于它。如果您将自己的硬件配置为具有电池，则修改您的Raft（或k / v服务器）以利用电池是有意义的。

问：本文提到硬件的使用，如DRAM和锂离子电池，以获得更好的恢复时间和存储空间。没有硬件优化的FaRM的使用是否仍然比以往的交易系统有任何好处？

A：我不确定FaRM是否会在没有非易失性RAM的情况下工作，因为单面日志写入（例如图4中的COMMIT-BACKUP）在电源故障时不会持续。您可以修改FaRM，以便在返回之前将所有日志更新都写入SSD，但会降低性能。SSD写入大约需要100微秒，而FaRM的单边RDMA写入非易失性RAM只需要几微秒。

问：我注意到这个系统的不稳定性依赖于在停电的情况下将DRAM复制到SSD的事实。为什么他们指定SSD，不允许复制到普通硬盘？他们只是试图使用最快的永久存储设备，因此没有简单的谈论硬盘 - 为什么？

答：是的，他们使用SSD，因为它们很快。他们可以在不改变设计的情况下使用硬盘。然而，在停电期间将数据写入磁盘将需要更长的时间，这将需要更大的电池。也许他们通过使用硬盘驱动器可以节省的钱会被电池成本上涨所夸大。

问：如果FaRM应用于低延迟高吞吐量存储，那么为什么要使用内存中的解决方案而不是常规的持久存储（因为主要的权衡是正确的 - 预先存储的是低延迟高吞吐量）？

A：RAM比硬盘或SSD有更高的吞吐量和更低的延迟。因此，如果您想要非常高的吞吐量或低延迟，则需要操作存储在RAM中的数据，并且在常见情况下必须避免使用磁盘或SSD。这就是FaRM的工作原理 - 通常所有数据都使用RAM，如果电源出现故障，只能使用SSD来保存RAM内容，并在电源恢复时将其还原。

问：我对FaRM论文中的RMDA如何运作感到困惑，或者它对于设计至关重要。是否正确的是，纸张给出的是针对CPU使用进行优化的设计（因为他们提到他们的设计主要是CPU限制）？所以即使我们使用SSD而不是内存中的访问，这将会起作用？

A：以前的系统一般受到网络或磁盘（或SSD）的限制。FaRM使用RDMA消除（或大幅放宽）网络瓶颈，并使用电池备份的RAM避免不必使用磁盘，从而将磁盘/ SSD作为瓶颈。但是总是有一些*资源可以防止性能无限。在FaRM的情况下，它是CPU速度。如果大多数FaRM访问必须使用SSD而不是内存，则FaRM将运行得更慢，并且将受SSD限制。

问：FaRM中的初级，备份和配置管理器之间的区别是什么？为什么有三个角色？

答：数据是在许多主/备份集之间划分的。备份的要点是存储分片的副本，以防主要出现故障。主要对分片中的数据执行所有读取和写入操作，而备份只执行写入（以保持其数据副本与主副本相同）。只有一个配置管理器。它跟踪哪些原色和备份是活着的，并跟踪数据如何在其间分割。在高层次上，这种安排类似于GFS，它也在许多主/备份集之间划分了数据，并且还拥有一个跟踪数据存储位置的主站。

问：FaRM在一系列尺度上是有用还是有效？作者描述了一种为PETABYTE存储运行$ 12 + / GB的系统，更不用说2000系统和网络基础架构支持所有DRAM + SSD + UPS的开销。对于他们适度地描述为“足以容纳许多有趣的应用程序的数据集”，这似乎很愚蠢 - 昂贵。你甚至如何测试和验证这样的东西？

答：我认为如果您需要每秒支持大量的交易，FaRM才是有趣的。如果您每秒只需要几千次交易，就可以使用像MySQL这样的成熟的技术。你可能设置一个比作者的90机器系统相当小的FaRM系统。但是，除非您正在分片和复制数据，否则FaRM无效，这意味着您需要至少四个数据服务器（两个分片，每个分片两个服务器）以及ZooKeeper的几台机器（尽管可能您可以在四台机器上运行ZooKeeper ）。那么也许你有一个系统的成本在$ 10,000美元的价格，可以执行每秒几百万简单的交易，这是相当不错的。

问：对于严格的可串行化性和Farm不能确保读取的原子性（即使提交的事务是可序列化的），我也有些困惑。

A：Farm只保证提交的事务的可序列化。如果交易看到他们正在谈论的那种不一致，FaRM将中止交易。应用程序必须处理不一致的意思，即它们不应该崩溃，以便他们可以得到尽可能多的请求提交，以便FaRM可以中止它们。

问：通过绕过内核，FaRM如何确保RDMA完成的读取是一致的？如果在交易中完成阅读，会发生什么？

一个; 这里有两个危险。首先，对于一个大的对象，读写器可以在并发事务写入之前读取对象的前半部分，并且在并发事务写入后的下半部分，这可能导致读取程序崩溃。其次，如果并发可写事务可能无法序列化，则读取事务不能被允许提交。

基于我对作者以前的NSDI 2014文章的阅读，第一个问题的解决方案是每个对象的每个缓存行都有版本号，单缓存行RDMA读写是原子的。读取事务的FaRM库将获取所有对象的缓存行，然后检查它们是否都具有相同的版本号。如果是，库将该对象的副本提供给应用程序; 如果没有，图书馆再次通过RDMA读取。第二个问题是通过第4节中描述的FaRM验证方案来解决。在VALIDATE步骤中，如果另一个事务已经写入了我们的事务开始读取的对象，那么我们的事务将被中止。

问：截断如何工作？什么时候可以删除日志条目？如果一个条目被截断的调用删除，所有以前的条目是否也被删除？

答：TC告诉初级和备份删除事务的日志条目，TC在其日志中看到所有的条目都有一个COMMIT-PRIMARY或COMMIT-BACKUP。为了恢复将知道事务完成，尽管截断，第62页提到，即使在截断之后，原始文件也记住已完成的事务ID。我一般不认为截断一个记录意味着截断所有以前的记录。可能是这样，每个TC按顺序截断自己的交易。由于每个主/备份每个TC具有单独的日志，结果可能是每个日志截断以日志顺序发生。

问：我有什么可以在COMMIT-BACKUP阶段中止。这只是由于硬件故障吗？

A：我相信 如果其中一个备份没有响应，并且TC崩溃，那么在恢复期间可能会中止该事务。

问：既然这是一个乐观的协议，那么当少量的资源需求很高时，它会受到影响吗？如果每次交易必须中止时，这个计划似乎要有效地回溯，这个计划似乎表现不佳。

A：是的，如果有很多冲突交易，FaRM似乎因为中止而表现不佳。

问：这个系统没有锁定性能的主要ham绳？

答：如果有很多冲突的交易，那么图4提交协议可能会导致大量的浪费中止。特别是因为交易会遇到锁。另一方面，对于作者测量的应用程序，尽管有锁，FaRM也取得了不俗的表现。很可能的一个原因是他们的应用程序具有相对较少的冲突交易，因此没有多少争用锁。

问：图7显示，当操作次数超过120M时，性能显着下降。是因为乐观的并发协议，所以经过一些门槛太多的交易中止？

答：我怀疑服务器每秒只能处理大约1.5亿次操作的限制。如果客户发送的操作比这更快，其中一些将不得不等待; 这种等待导致延迟增加。

问：如果是，那么有什么方法a）使表现图表更平坦？b）避免在120M +范围内？

答：由用户确定服务器在发送请求之前没有超载？通常这些系统有一个自我调节的效果 - 如果系统速度较慢（延迟时间很长），客户端会产生更慢的请求。由于每个客户端都会生成一系列请求，所以如果第一个请求被延迟，则会延迟客户端提交下一个请求。或者因为最终用户导致请求被生成，并且用户感到无聊等待缓慢的系统并且消失。但是，无论谁部署系统，都必须小心确保存储系统（以及系统的所有其他部分）足够快以便以相当低的延迟执行可能的工作负载。