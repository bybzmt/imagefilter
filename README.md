#imagefilter

图片过滤处理，将图片处理成指定样式输出

参数 | 说明     | 类型
-----|----------|------------
o    | 动作类型 | 可选项: ori, resize, crop, fit, fill
w    | 宽       | int 如果为0则为原图宽
h    | 高       | int 如果为0则为原图高
f    | 输出格式 | 默认为自动, 支持 jpg,png,gif
t    | 随机串   | string
a    | 锚点     | 可选项: center, topleft, top, topright, left, right, bottomleft, bottom, bottomright
s    | 签名     | base64 URLEncoding



动作   | 说明
-------|-------
ori    | 原图, 此时宽高不生效
resize | 缩放
crop   | 裁切
fit    | 等比例缩放
fill   | 等比例缩放，并冲满整图，切掉多余的部分

这个程序是用在用户记问时生成相应尺寸的图片，这样子可以在上传图片时省掉很多操作。
此前程序必需架在nginx的后端，并且nginx必需开启缓存！
