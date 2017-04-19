<?php

require __DIR__ . "/imagefilter.php";

bybzmt\imagefilter::$signatureKey = "abcd";

$image = new bybzmt\imagefilter();

//指定格式
//$url = $image->build_url("/imgs/1.png", 'fill', 300, 300);
$url = $image->build_url("/imgs/1.png", 'fill', 300, 300, 'jpg');
var_dump($url);
//解码
var_dump($image->decode($url));

