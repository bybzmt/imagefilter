<?php

require __DIR__ . "/imagefilter.php";

bybzmt\imagefilter::$signatureKey = "abcd";

$image = new bybzmt\imagefilter();

$url = $image->build_url("/imgs/1.png", 'fill', 300, 300);
var_dump($url);
//解码
var_dump($image->decode($url));

//指定格式
//var_dump($image->build_url("/imgs/1.png", 'fit', 300, 300, 'jpeg'));
