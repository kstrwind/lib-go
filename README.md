# lib-go
常用的Lib库

## kdebug
debug 工具

### VarDump
类似PHP的var_dump函数的debug函数，用于输出当前变量的类型及值， 特点如下：

1. 递归地输出变量的所有成员类型及值
2. 采用4空格来缩进，同一缩进级别表示同一层次的变量
3. 对于指针，会在同一行解引用，显示指针指向的数据结构
4. 对于数组、切片、映射、字符串等类型，会显示其长度
5. 在切片、数组、映射结尾，会有结束标记 `//end $type` 

*特殊标记说明：*

- -->, 表示指针解引用
- ==>, 表示数组、切片、映射、字符串等类型的值
- :, 表示纯量变量或者结构体的成员等
- (), 注释数组、切片、映射、字符串的长度

*使用方法*
```
var cfg *Conf = &Conf{}
err = yaml.Unmarshal(data, cfg)
if err != nil {
    return err
}
kdebug.VarDump(cfg)

输出：
[*main.Conf] --> struct[main.Conf] ==> {
    [*main.Person] --> struct[main.Person] ==> {
        Name [string](6) ==> "peter"
        Height [int] : 180
        Weight [int] : 100
    } //end struct
    slice[[]*main.Person](2) ==>
        [*main.Person] --> struct[main.Person] ==> {
            Name [string](2) ==> "p1"
            Height [int] : 111
            Weight [int] : 111
        } //end struct
        [*main.Person] --> struct[main.Person] ==> {
            Name [string](2) ==> "p2"
            Height [int] : 222
            Weight [int] : 222
        } //end struct
    //end slice
} //end struct
```
