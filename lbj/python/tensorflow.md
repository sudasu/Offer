# tensorflow https://blog.csdn.net/kangshuangzhu/article/details/106851826

## tfrecord

创建tfRecord的example实例:

```python

value_city = u"北京".encode('utf-8')   # 城市
value_use_day = 7                      # 最近7天打开淘宝次数
value_pay = 289.4                      # 最近7天消费金额
value_poi = [b"123", b"456", b"789"]   # 最近7天浏览电铺

'''
下面生成ByteList,Int64List和FloatList
'''
bl_city = tf.train.BytesList(value = [value_city])  ## tf.train.ByteList入参是list,所以要转为list
il_use_day = tf.train.Int64List(value = [value_use_day])
fl_pay = tf.train.FloatList(value = [value_pay])
bl_poi = tf.train.BytesList(value = value_poi)

'''
下面生成tf.train.Feature
'''
feature_city = tf.train.Feature(bytes_list = bl_city)
feature_use_day = tf.train.Feature(int64_list = il_use_day)
feature_pay = tf.train.Feature(float_list = fl_pay)
feature_poi = tf.train.Feature(bytes_list = bl_poi)
'''
下面定义tf.train.Features
'''
feature_dict = {"city":feature_city,"use_day":feature_use_day,"pay":feature_pay,"poi":feature_poi}
features = tf.train.Features(feature = feature_dict)
'''
下面定义tf.train.example
'''
example = tf.train.Example(features = features)
print(example)
```

输出结果

```python
features {
  feature {
    key: "city"
    value {
      bytes_list {
        value: "\345\214\227\344\272\254"
      }
    }
  }
  feature {
    key: "pay"
    value {
      float_list {
        value: 289.3999938964844
      }
    }
  }
  feature {
    key: "poi"
    value {
      bytes_list {
        value: "123"
        value: "456"
        value: "789"
      }
    }
  }
  feature {
    key: "use_day"
    value {
      int64_list {
        value: 7
      }
    }
  }
}
```
