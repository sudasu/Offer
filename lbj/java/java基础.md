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

连接就是将一个二进制格式的类或结构转换成jvm中的运行时可以使用的状态。为了连接时的灵活性(如遇见递归加载的情况)，在遵循java语言的规范---(1.初始化前类或接口必须被完全验证和准备2.在连接期间检测到的程序需要连接包含错误的类和接口时，抛出相关错误。)的情况下，选择懒解析的方式对单独正在被使用的符号引用进行解析。

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