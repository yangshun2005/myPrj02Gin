#### 阅读源码记录
`主要记叙阅读后段自写的代码`

1.  首页  get   "/"
1.1 获取数据：   posts  total   policy  user    

1.2 逻辑：   返回template： index.html

1.3 models:     
```text
ListPublishedPost()     
CountPostByTag()    
ListTagByPostId()   
```

1.4 controller: 
```text
bluemonday.StrictPolicy()
system.GetConfiguration()

```

1.5 extend-controller：
```text
c.AbortWithStatus()
bluemonday.StrictPolicy()
strconv.FormatUint()
data.sql
DB.Rows()
```