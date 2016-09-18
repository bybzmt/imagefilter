<?php
namespace bybzmt;

class imagefilter
{
	static public $signatureKey="";

	public function build_url($path, $op, $width, $height, $format="", $anchor="")
	{
		$op = $this->_val_op($op);
		$format = $this->_val_format($format);
		$anchor = $this->_val_anchor($anchor);

		// data(6byte+) = op(4bit) + anchor(4bit) + format(1byte) + width(2byte) + height(2byte) + path
		$data = pack('C', $op << 4 | $anchor) . pack('C', $format) . pack('n', $width) . pack('n', $height) . $path;

		if (self::$signatureKey) {
			$sign = hash_hmac("md5", $data, self::$signatureKey, true);
			//sign len(1byte) + sign
			$sign = pack('C', strlen($sign)) . $sign;
		} else {
			$sign = pack('C', 0);
		}

		//final(9byte+) = protol verion(1byte) + sign + data
		$final = pack('C', 1) . $sign . $data;

		$url = '/'.$this->base64url_encode($final);
		return $url;
	}

	private function base64url_encode($data)
	{
		return rtrim(strtr(base64_encode($data), '+/', '-_'), '=');
	}

	private function _val_op($op)
	{
		switch (strtolower($op)) {
		case 'ori': return 1;
		case 'resize': return 2;
		case 'crop': return 3;
		case 'fit': return 4;
		case 'fill': return 5;
		default: throw new \Exception("undefined op: {$op}");
		}
	}

	private function _val_format($format)
	{
		switch (strtolower($format)) {
		case '': return 0;
		case 'jpg': return 1;
		case 'jpeg': return 1;
		case 'png': return 2;
		case 'gif': return 3;
		default: throw new \Exception("undefined format: {$format}");
		}
	}

	private function _val_anchor($anchor)
	{
		switch (strtolower($anchor)) {
		case 'topleft': return 1;
		case 'top': return 2;
		case 'topright': return 3;
		case 'left': return 4;
		case 'center': return 5;
		case '': return 5;
		case 'right': return 6;
		case 'bottomleft': return 7;
		case 'bottom': return 8;
		case 'bottomright': return 9;
		default: throw new \Exception("undefined anchor: {$anchor}");
		}
	}
}
