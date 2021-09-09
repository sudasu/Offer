# [java](https://docs.oracle.com/en/java/index.html)([java8](https://docs.oracle.com/javase/8/))

## [java规范引用](https://docs.oracle.com/javase/specs/jls/se8/html/index.html)

## [java specification](https://docs.oracle.com/javase/specs/index.html)

## classloader

类加载流程:加载->连接(验证->准备->解析)->初始化。可以通过使用classLoader的loadClass来显示加载类，也可以使用new或者调用该类的静态方法来隐式加载，new的实现流程暂不清楚，但是可以清楚classLoader不会立即解析而隐式加载会。根据猜测，每个类会引用自己的类加载器，那么这个类在调用加载其他类信息是隐式加载是采用自己的类加载器？根据双亲委派机制，子加载器加载的类对于父加载器是可见的，此设计应该是在findLoadedClass时直接将子加载器加载的类返回。

### getClassLoader()

每一个class对象都会有引用指向对应classloader，可通过class.getClassLoader()方法调用，但如果是bootstrap classloader则会返回Null。对于Array对象，getClassLoader()返回的是array元素的类加载器，如果是初始类型则不会返回类加载器。

### getContextClassLoader()

java Thread中拥有getContextClassLoader()方法去获取上下文类加载器，也可使用set方法去设置其他类加载器，其中Thread的当前线程实例获取的classLoader默认为AppClassLoader。

### loadClass()

loadClass(String name)是public方法，在调用时会调用protected方法loadClass(String name,boolean resolve)函数，该调用是不解析的。也就是说，并不会执行static代码块，根据测试确实是如此。而protected的loadClass一般做三步操作，先查找虚拟机内是否已加载该类信息，然后调用父类加载，如果都不行则最后自己调用findClass()来查找加载(一般自定义的类加载器都是重写该方法)。但是值得注意的一点是在loadClass方法内部，如果是并发加载则会有concurrentHashMap维护的哈希表来保证一个className一个锁，对锁的粒度进行细化增强并发。否则将会对该classLoader对象进行加锁，效率不高。还有需要注意的是，避免在加载两个类时的循环依赖，如果执行顺序不正确会有产生死锁的风险。(不知道为什么，在调试时f5进入失效，只能断点调试，因为子线程的原因？)

### 载入自定义java.\*

我们自己实现的java.\*包下的类是无法被类装载器所加载的，一方面是双亲委派，另一方面jvm的实现也确保了java.\*目录下的class不会被除bootstrap类加载器以外加载器加载。