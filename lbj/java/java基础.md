# [java](https://docs.oracle.com/en/java/index.html)([java8](https://docs.oracle.com/javase/8/))

## [java语言与虚拟机规范](https://docs.oracle.com/javase/specs/index.html)

### [java8 语言规范](https://docs.oracle.com/javase/specs/jls/se8/html/index.html)

### [java8 虚拟机规范](https://docs.oracle.com/javase/specs/jvms/se8/html/index.html)

## package

### 顶级类

如果一个顶级类缺少了访问修饰符，则只能被本包的编译单元访问。为了使得其他包的代码可以访问，需要将访问修饰符设为public。顶级类不能使用static,protected,private修饰符，可以使用abstract,public,final。

## 类加载

### 概述

java虚拟机通过加载某个指定类，并调用该类的main方法来启动的，类加载流程主要还是由:`加载->连接(验证->准备->解析)->初始化`三个步骤构成。

### 加载

所谓加载就是指通过特定的名称找到二进制表示的class或者interface，并据此构造一个二进制的class对象表示该class或interface。二进制文件的生成，一般情况是通过Java编译器提前编译源码转换成二进制文件，但也可以动态计算转换。其中class file的格式由jvm定义，但是也可以通过类加载器的defineClass方法使用其他格式的二进制文件构造。

好的加载器需要注意两点：1.对于相同的名称，类加载器应该只返回相同的类对象。2.对于L1委托L2加载的情况，在任何返回类对象的场景，L1和L2应该返回相同得到类对象。(当然如果这些原则被违规的类加载器打破了，虚拟机会保证类型安全系统不被破坏。)

不同的类加载器采用不同的加载策略，有些性能优化的选择，比如组加载和预测预加载。但是在使用这些选择时，有些错误可能无法被立刻探测到，比如老的类对象在缓存而实际该对象已经被删了应该抛出ClassNotFound的错误。类加载错误抛出的error为LinkageError的子类，ClassCircularityError(自己是自己的父类)，ClassFormatError(格式不符合要求)，NoClassDefFoundError(类信息未找到)。

### 连接

连接就是将一个二进制格式的类或结构转换成jvm中的运行时可以使用的状态。为了连接时的灵活性(如遇见递归加载的情况)，在遵循java语言的规范---(1.初始化前类或接口必须被完全验证和准备2.在连接期间检测到的程序需要连接包含错误的类和接口时，抛出相关错误。)的情况下，选择懒解析的方式对单独正在被使用的符号引用进行解析。注意:其中classLoader的native方法resolveClass执行的是连接操作。

#### 验证

验证是为了保证加载的二进制类信息文件结构正确。一般包括是否每行指令都有相应的操作代码，跳转指令是否正确跳转至指令的开始部分，所有指令是否遵循jvm的类型规则，每个方法是否都结构正确。在验证错误发生时，会抛出LinkageError的子类，VerifyError。

#### 准备

为class的静态变量，常量分配内存空间设置初始化值，其中常量此时赋值。为了使后面的操作更加高效，jvm此时也做了些额外的操作，如维护一个方法表以保证该类实例被调用时避免去父类搜索。(是相当于自己维护一个方法表，将父类的方法引用放入？)

#### 解析

二进制表示的class或者interface通过二进制名称符号引用他们的属性，方法，构造。为了验证引用正确性以及后续重复使用的高效，需要对其进行解析转换成直接引用。解析过程会抛出IncompatibleClassChangeError(LinkageErro子类)或其子类Error，如：

* IllegalAccessError：不合规范的访问被private,protected等修饰的变量。(这发生在引用public字段编译成功后，被引用类对该字段进行了修改。)
* InstantiationError：实例化的abstract修饰的类，同样可能发生在编译成功后的被引用类修改。
* NoSuchFieldError：不存在该字段，错误发生同上。
* NoSuchMethodError：不存在该方法，错误发生同上。

除此之外LinkageError的子类UnsatisfiedLinkError，发生在native方法找不到实现时抛出。

### 初始化

初始化一般就是对静态变量进行赋值，以及对static代码段进行执行。初始化的触发：1.class T被实例化。2.class T的静态方法被调用。3.class T的静态字段被使用且不是常量。4.class T的静态字段被赋值。(注意：编译器可能会对接口合成一些默认方法，这些方法既没被显示也没被隐式的声明。这些方法可能触发接口的初始化，哪怕源代码并没有指示该接口需要被初始化？)为了实现class或者接口的使用具有一致状态的初始化器的目的，该初始化器遵守维持第一次被观察到类型状态的规则。静态变量和类变量初始化器按照文本顺序执行，对于使用接口中的变量，如果是常量则该类不会被初始化，而如果是表达式则会导致该接口初始化。

### 双亲委派

双亲委派，向上可以保证共享已加载的类信息，向下可以保证同全限定名不同类加载器的隔离，避免冲突。

可以通过使用classLoader的loadClass来显式加载类，也可以使用new或者调用该类的静态方法来隐式加载，new的实现流程暂不清楚，但是可以清楚classLoader不会立即解析而隐式加载会。根据猜测，每个类会引用自己的类加载器，那么这个类在调用加载其他类信息是隐式加载是采用自己的类加载器？根据双亲委派机制，子加载器加载的类对于父加载器是可见的，此设计应该是在findLoadedClass时直接将子加载器加载的类返回。

### getClassLoader()

每一个class对象都会有引用指向对应classloader，可通过class.getClassLoader()方法调用，但如果是bootstrap classloader则会返回Null。对于Array对象，getClassLoader()返回的是array元素的类加载器，如果是初始类型则不会返回类加载器。

### getContextClassLoader()

java Thread中拥有getContextClassLoader()方法去获取上下文类加载器，也可使用set方法去设置其他类加载器，其中Thread的当前线程实例获取的classLoader默认为AppClassLoader。

### loadClass()

loadClass(String name)是public方法，在调用时会调用protected方法loadClass(String name,boolean resolve)函数，该调用是不解析的。也就是说，并不会执行static代码块，根据测试确实是如此。而protected的loadClass一般做三步操作，先查找虚拟机内是否已加载该类信息，然后调用父类加载，如果都不行则最后自己调用findClass()来查找加载(一般自定义的类加载器都是重写该方法)。但是值得注意的一点是在loadClass方法内部，如果是并发加载则会有concurrentHashMap维护的哈希表来保证一个className一个锁，对锁的粒度进行细化增强并发。否则将会对该classLoader对象进行加锁，效率不高。还有需要注意的是，避免在加载两个类时的循环依赖，如果执行顺序不正确会有产生死锁的风险。(不知道为什么，在调试时f5进入失效，只能断点调试，因为子线程的原因？)

### 载入自定义java.\*

我们自己实现的java.\*包下的类是无法被类装载器所加载的，一方面是双亲委派，另一方面jvm的实现也确保了java.\*目录下的class不会被除bootstrap类加载器以外加载器加载。

## 线程和锁

## JNDI

Java Naming and Directory Interface,是java命名和目录接口。是一个应用程序设计的API，为开发人员提供查找和访问各种**命名和目录**服务的通用、统一的接口，类似JDBC都是构建在抽象层上。JNDI中的命名就是将Java对象以某个名称的形式绑定到一个容器环境中，如将数据源对象绑定到JNDI环境中，后续的Servlet程序就可以直接从JNDI环境中查询出这个对象进行使用。容器环境本身也是一个Java对象，容器环境可以绑定到另一个Context对象中，形成树级结构。命名与目录两者之间的关键差别是目录服务中对象不但可以有名称还可以有属性（例如，用户有email地址），而命名服务中对象没有属性。

### Context

Context由一系列 name-to-object bindings组成，核心功能为查询，绑定，解绑，重命名对象，以及创建和销毁子环境。

### Names

每一个在Context接口中的Nameing相关方法，通常都有两种重载的实现，参数分别为Name类型和string类型。Naming对象对应用来说提供了许多方便操作，包括组合修改name，比较name组成部分等。

### Binding

Binding类表示一个name-to-object对象，是一个包含绑定对象的name,class name,以及自己本身的元组。Binding类是NameClassPair类的子类，NameClassPair类描述对象和对象类信息。

### References

对象以不同的方式存储在naming and directory services里面，比如如果支持存储对象的话就可以以序列化的方式进行存储。但是，有些服务可能是不支持这种方式的，毕竟对于目录中的对象java程序也仅仅是访问该服务的一类程序。JNDI定义了reference,用来表示如果构造该对象副本的信息。JNDI将会尝试将从directory查询出的引用转换成Java对象，以至于使JNDI的客户端认为存储在目录中的是java object。

### Initial Context

在JNDI中，所有的naming and directory操作都是基于相对上下文环境所执行的，没有一个明确的root。因此，JNDI定义了一个初始环境--InitialContext,作为naming and directory操作的起点。

### Directory Context

该环境定义了检查和更新directory object或directroy enty属性的方法，主要用作提供dierctory services。Directroy Context实现了Context接口，同时也可以提供naming services。

### Naming Event

该Event用于反映在naming/directory服务中产生的事件，由Context和DirContext的子接口EventDirContext定义。NamingEvent代表的事件触发有两种类型：1.namespace相关的影响(add/remove/rename一个对象)。2.对象类容的修改。每类事件的处理由相应的监听器来维护：NamespaceChangeListener,ObjectChangeListener。代码例子：

```
EventContext src = 
    (EventContext)(new InitialContext()).lookup("o=wiz,c=us");
src.addNamingListener("ou=users", EventContext.ONELEVEL_SCOPE,
    new ChangeHandler());
...
class ChangeHandler implements ObjectChangeListener {
    public void objectChanged(NamingEvent evt) {
        System.out.println(evt.getNewBinding());
    }
    public void namingExceptionThrown(NamingExceptionEvent evt) {
        System.out.println(evt.getException());
    }
}
```

注意：context相关操作需要保证自己保证线程安全，如为环境增加监听器。

## NIO

### ByteBuffer

java NIO引入了三种类型的ByteBuffer：1.HeapByteBuffer，通过ByteBuffer.allocate方法创建，分配JVM堆空间，可以获得GC支持，缓存优化。但是由于不是页面对齐的，所以如果需要通过JNI与本地代码进行交互，JVM会复制到对齐的缓冲区空间。2.DirectByteBuffer，通过ByteBuffer,allocateDirect方法创建，分配JVM外的堆外空间。由于不是JVM管理的，所以内存空间是页面对齐的不受GC影响，是处理本地代码的最好选择，但是必须自己管理这块内存防止内存泄漏。3.MappedByteBuffer，通过FileChannel.map创建，也是不受JVM管理的堆外空间，但不同的是作为OS mmap系统调用的包装。

## Stream

### 常用方法

```java

//流的转换操作api

Stream<T> filter(Predicate<? super T> predicates)                                    //例(w->w.length()<10)，用于筛选满足条件的结果，其中Predicate指返回bool的Function

<R> Stream<R> map(Function<? super T,? extends R> mapper)                            //例(String::toLowerCase);(s->s.substring(0,1)),其中T为入参，R为回参，用于对流中的每个元素都进行操作。

<R> Stream<R> map(Function<? super T,? extends Stream<? extends R>> mapper)          //与上的区别在于对返回的参数做了限制，即返回的是流，正常情况下是流的流，但该函数将会将合并成单个流。

Stream<T> limit(long maxSize)                                                        //返回当前流的前maxSize个元素的流

Stream<T> skip(long n)                                                               //返回跳过前n个元素的所有元素的流

Stream<T> peek(Consumer<? super T> action)                                           //对当前元素进行操作，但不影响流，其中Consumer是返回值为void的Function

static Stream<T> concat(Stream<? extends T> a,Stream<? extends T> b)                 //产生a流后接b流的合并流

//其他筛选如distinct去重，sorted排序

//流的终结操作包含 max,min,findFirst,findAny,anyMatch,allMatch,noneMatch。其中find相关操作返回是Optional对象，Optional包含了许多判空，及提供默认值等相应处理函数。



optionV.f().g() => optionV.f().flatMap(T::g);

<U> Optional<U> flatMap<Function<? super T,Optional<U>> mapper>                      //同一个类的某些计算或者函数转换成对Optional的封装，由于f的转换返回Optional类型或者直接就是Optional类型无法链式，所以使用flatMap封装。注意：Function的入参






```
