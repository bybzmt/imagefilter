<?php

require __DIR__ . "/imagefilter.php";

bybzmt\imagefilter::$signatureKey = "abcd";

$url = new bybzmt\imagefilter();

var_dump($url->build_url("/imgs/1.png", 'fit', 300, 300));

//指定格式
var_dump($url->build_url("/imgs/1.png", 'fit', 300, 300, 'jpeg'));
