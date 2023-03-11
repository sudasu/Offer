# java

## 集合

### list

#### ArrayList

1. 对于有长度限制的ArrayList,扩容遵循1.5倍扩容的原则即length+length >> 1,而Vector是两倍扩容比较浪费空间。
2. ArrayList是非线程安全的，Vector是线程安全的但读和写都需要加锁，CopyOnWriteArrayList读写分离、对写加锁是线程安全的。

```java

    // Collections.synchronizedList
    // 并发限制的重要变量，通过对原List的所有方法synchronized(mutex){}操作，保证并发限制
    final Object mutex;     // Object on which to synchronize

    // CopyOnWriteArrayList的重要变量，其中采用可重入锁的方式保证写的时候加锁
    // 加锁后对volatile变量的数组对象进行写，然后替换时触发缓存一致性
    /** The lock protecting all mutators */
    final transient ReentrantLock lock = new ReentrantLock();

    /** The array, accessed only via getArray/setArray. */
    private transient volatile Object[] array;

    public boolean add(E e) {
        final ReentrantLock lock = this.lock;
        lock.lock(); //上锁，只允许一个线程进入
        try {
            Object[] elements = getArray(); // 获得当前数组对象
            int len = elements.length;
            Object[] newElements = Arrays.copyOf(elements, len + 1);//拷贝到一个新的数组中
            newElements[len] = e;//插入数据元素
            setArray(newElements);//将新的数组对象设置回去
            return true;
        } finally {
            lock.unlock();//释放锁
        }
    }
```
#### LinkedList

一般List的使用都是ArrayList,因为除了扩容麻烦一些外，大部分时候数组实现都比链表实现性能高，特别是随机查找方面。链表一般还是用作队列比较多，如ConcurrentLinkedQueue

```java
// 入队操作
public boolean offer(E e) {
    checkNotNull(e);
    // 入队前，创建一个入队节点
    final Node<E> newNode = new Node<E>(e);

    for (Node<E> t = tail, p = t;;) {
        // 创建一个指向tail节点的引用
        Node<E> q = p.next;
        if (q == null) {
            // p is last node
            if (p.casNext(null, newNode)) {
                // Successful CAS is the linearization point
                // for e to become an element of this queue,
                // and for newNode to become "live".
                if (p != t) // hop two nodes at a time
                    casTail(t, newNode);  // Failure is OK.
                return true;
            }
            // Lost CAS race to another thread; re-read next
        }
        else if (p == q)
            // We have fallen off list.  If tail is unchanged, it
            // will also be off-list, in which case we need to
            // jump to head, from which all live nodes are always
            // reachable.  Else the new tail is a better bet.
            p = (t != (t = tail)) ? t : head;
        else
            // Check for tail updates after two hops.
            p = (p != t && t != (t = tail)) ? t : q;
    }
}
```

### hashmap

hashmap为什么大小是2的倍数。

1. 判断简单，只需与-1值做或|算判断是否等于0即可
2. 取余简单，与-1值做且&运算能快速得到余值
3. 避免扩容迁移时的再碰撞，将Hashcode与原数组大小做&运算，为0和为1分别放置在不同位置，为1的放在1+原数组大小的桶。

## 并发

### synchornized

1. 对象锁：对obj进行synchornized关键字包含，或方法申明时加synchornized关键字
2. 类锁：对类的静态方法加synchornized关键字，或对obj.class使用关键字包含

特点：1.是可重入锁，但是是非公平的，唤醒阻塞线程是随机的。2.比较重量级，使用操作系统的阻塞实现，会引起系统调用。

obj的wait和notify方法：其中wait方法必须在synchornized代码块中使用，在调用后立即放弃持有对象锁，可通过其他持有对象锁的代码块使用notify方法唤醒。其中注意的是sleep睡眠是抱锁睡眠，不会进入等待队列而是进入阻塞状态。wait方法会立即释放锁中断，进入等待队列。唤醒后仍然需要去争抢获得同步锁才能继续从中断处向下执行。

## 垃圾回收

### 标记清除
它是最基础的收集算法。
原理：分为标记和清除两个阶段：首先标记出所有的需要回收的对象，在标记完成以后统一回收所有被标记的对象。
特点：（1）效率问题，标记和清除的效率都不高；（2）空间的问题，标记清除以后会产生大量不连续的空间碎片，空间碎片太多可能会导致程序运行过程需要分配较大的对象时候，无法找到足够连续内存而不得不提前触发一次垃圾收集。
地方 ：适合在老年代进行垃圾回收，比如CMS收集器就是采用该算法进行回收的。

### 标记整理
原理：分为标记和整理两个阶段：首先标记出所有需要回收的对象，让所有存活的对象都向一端移动，然后直接清理掉端边界以外的内存。
特点：不会产生空间碎片，但是整理会花一定的时间。
地方：适合老年代进行垃圾收集，parallel Old（针对parallel scanvange gc的） gc和Serial old收集器就是采用该算法进行回收的。

### 复制算法

原理：它先将可用的内存按容量划分为大小相同的两块，每次只是用其中的一块。当这块内存用完了，就将还存活着的对象复制到另一块上面，然后把已经使用过的内存空间一次清理掉。
特点：没有内存碎片，只要移动堆顶指针，按顺序分配内存即可。代价是将内存缩小位原来的一半。
地方：适合新生代区进行垃圾回收。serial new，parallel new和parallel scanvage
收集器，就是采用该算法进行回收的。
复制算法改进思路：由于新生代都是朝生夕死的，所以不需要1：1划分内存空间，可以将内存划分为一块较大的Eden和两块较小的Suvivor空间。每次使用Eden和其中一块Survivor。当回收的时候，将Eden和Survivor中还活着的对象一次性地复制到另一块Survivor空间上，最后清理掉Eden和刚才使用过的Suevivor空间。其中Eden和Suevivor的大小比例是8：1。缺点是需要老年代进行分配担保，如果第二块的Survovor空间不够的时候，需要对老年代进行垃圾回收，然后存储新生代的对象，这些新生代当然会直接进入来老年代。

### 优化收集方法

分代收集算法
原理：根据对象存活的周期的不同将内存划分为几块，然后再选择合适的收集算法。
一般是把java堆分成新生代和老年代，这样就可以根据各个年待的特点采用最适合的收集算法。在新生代中，每次垃圾收集都会有大量的对象死去，只有少量存活，所以选用复制算法。老年代因为对象存活率高，没有额外空间对他进行分配担保，所以一般采用标记整理或者标记清除算法进行回收。

### CMS垃圾回收过程

初始标记(stw)->并发标记->并发清理->重新标记->并发清理