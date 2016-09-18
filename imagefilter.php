<?php
namespace bybzmt;

class imagefilter
{
	static public $signatureKey="";

	public function build_url($path, $op, $width, $height, $format="", $anchor="")
	{
		$randstr = mt_rand(0, 2000000000);
		$msg = $path . $op . $width . $height . $format . $anchor . $randstr;

		$sign = self::base64url_encode(hash_hmac("md5", $msg, self::$signatureKey, true));

		$data = array(
			't' => $randstr,
			'o' => $op,
			'w' => $width,
			'h' => $height,
			's' => $sign,
		);
		if ($format) {
			$data['f'] = $format;
		}
		if ($anchor) {
			$data['a'] = $anchor;
		}

		return $path . "?" . http_build_query($data);
	}

	public function base64url_encode($data)
	{
		return rtrim(strtr(base64_encode($data), '+/', '-_'), '=');
	}
}
