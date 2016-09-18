#imagefilter

图片过滤处理，将图片处理成指定样式输出

这个程序在用户访问时生成相应尺寸的图片，这样子可以在上传图片时省掉很多操作。

程序本身没有缓存,它必需架在nginx的后端，并且nginx必需开启缓存！

编译: `go build imagefilter.go` 后直接复制需要用的地方即可。

#php使用

安装 : `composer bybzmt/imagefilter`

```php
public function build_url($path, $op, $width, $height, $format="", $anchor=""){}
```

参数   | 说明
-----  | -------------
path   | 图片原始路径
op     | 动作类型。可选项见下表
width  | 输出图片的宽。 如果为0则为原图宽
height | 输出图片的高。 如果为0则为原图高
format | 输出格式, 默认为原图格式, 支持 jpg,png,gif
anchor | 锚点,裁切时定位。见锚点位置说明


动作类型表：

动作   | 说明
-------|-------
ori    | 原图, 此时宽高不生效
resize | 缩放
crop   | 裁切
fit    | 等比例缩放
fill   | 等比例缩放，并冲满整图，切掉多余的部分

锚点位置说明：

位置 | 左         | 中     | 右
-----|------------|--------|---------
上   | topLeft    | top    | topRight
中   | left       | center | right
下   | bottomLeft | bottom | bottomRight


