<?php
namespace bybzmt;

class imagefilter
{
	static public $signatureKey="";

	public function build_url($path, $op, $width, $height, $format="", $anchor)
	{
		$rand = mt_rand(0, PHP_INT_MAX);
		$msg = $path . $op . $width . $height . $format . $anchor . $randstr;

		$sign = self::base64url_encode(hash_hmac("sha256", $msg, self::$signatureKey, true));

		$data = array(
			't' => $randstr,
			'o' => $op,
			'w' => $width,
			'h' => $height,
			'f' => $format,
			'a' => $anchor,
			's' => $sign,
		);

		return $path . "?" . http_build_query($data);
	}

	public function base64url_encode($data)
	{
		return rtrim(strtr(base64_encode($data), '+/', '-_'), '=');
	}
}
