<?php
namespace bybzmt;

class imagefilter
{
	static public $signatureKey="";

	public function resize($path, $src_width, $src_height, $dst_width, $dst_height, $format="jpeg_gif")
	{
		$req = $this->_build_resize($path, $src_width, $src_height, $dst_width, $dst_height, $format);
		$sign = $this->_build_sign($req);
		return $this->_encode_url($sign, $req);
	}

	public function ori_url($path, $format="jpeg_gif")
	{
		$req = $this->_build_ori($path, $format);
		$sign = $this->_build_sign($req);
		return $this->_encode_url($sign, $req);
	}

	public function fit_url($path, $width, $height, $format="jpeg_gif")
	{
		$req = $this->_build_fit($path, $width, $height, $format);
		$sign = $this->_build_sign($req);
		return $this->_encode_url($sign, $req);
	}

	public function fill_url($path, $width, $height, $anchor,  $format="jpeg_gif")
	{
		$req = $this->_build_fill($path, $width, $height, $anchor,  $format);
		$sign = $this->_build_sign($req);
		return $this->_encode_url($sign, $req);
	}

	public function build_url($path, $op, $width, $height, $format="", $anchor="")
	{
		$op = $this->_val_op($op);
		$format = $this->_val_format($format);
		$anchor = $this->_val_anchor($anchor);

		if ($format) {
			$s_format = ".".$this->_str_format($format);
		} else {
			$s_format = "";
		}

		// data(6byte+) = op(4bit) + anchor(4bit) + format(1byte) + width(2byte) + height(2byte) + path
		$data = pack('C', $op << 4 | $anchor) . pack('C', $format) . pack('n', $width) . pack('n', $height) . $path;

		if (self::$signatureKey) {
			$sign = hash_hmac("md5", $data . $s_format, self::$signatureKey, true);
			//sign len(1byte) + sign
			$sign = pack('C', strlen($sign)) . $sign;
		} else {
			$sign = pack('C', 0);
		}

		//final(8byte+) = protol verion(1byte) + sign + data
		$final = pack('C', 1) . $sign . $data;

		$url = '/'.$this->_base64url_encode($final) . $s_format;
		return $url;
	}

	public function decode($url)
	{
		if (strlen($url) < 9) {
			return false;
		}

		$idx = strpos($url, '.');
		if ($idx !== false) {
			$url = substr($url, 0, $idx);
		}

		$data = $this->_base64url_decode(substr($url, 1));

		//var_dump(unpack('C*', $data));
		$sign_len = unpack('C', $data[1])[1];
		if (strlen($data) < $sign_len + 8) {
			return false;
		}

		$sign = substr($data, 2, $sign_len);
		$params = substr($data, 2+$sign_len, 6);
		$path = substr($data, 2+$sign_len+6);

		$tmp = unpack('C*', $params);
		$op = $tmp[1] >> 4;
		$anchor = $tmp[1] & 0xf;
		$format = $tmp[2];
		$width = $tmp[3]<<8 | $tmp[4];
		$height = $tmp[5]<<8 | $tmp[4];

		return array(
			'op' => $this->_str_op($op),
			'anchor' => $this->_str_anchor($anchor),
			'format' => $this->_str_format($format),
			'width' => $width,
			'height' => $height,
			'path' => $path,
			'sign' => $sign,
		);
	}

	private function _build_sign($data)
	{
		if (self::$signatureKey) {
			$sign = hash_hmac("md5", $data, self::$signatureKey, true);
			//sign len(1byte) + sign
			$sign = pack('C', strlen($sign)) . $sign;
		} else {
			$sign = pack('C', 0);
		}

		return $sign;
	}

	private function _encode_url($sign, $data)
	{
		$final = pack('C', 1) . $sign . $data;

		$url = '/'.$this->_base64url_encode($final);
		return $url;
	}

	private function _build_ori($path, $format="jpeg_gif")
	{
		$op = $this->_val_op('ori');
		$format = $this->_val_format($format);

		return pack('C', $op) .
			pack('C', $format) .
			$path;
	}

	private function _build_resize($path, $src_width, $src_height, $dst_width, $dst_height, $format="jpeg_gif")
	{
		$op = $this->_val_op('resize');
		$format = $this->_val_format($format);
		$anchor = $this->_val_anchor($anchor);

		return pack('C', $op) .
			pack('n', $src_width) .
			pack('n', $src_height) .
			pack('n', $dst_width) .
			pack('n', $dst_height) .
			pack('C', $format) .
			$path;
	}

	private function _build_fit($path, $width, $height, $format="jpeg_gif")
	{
		$op = $this->_val_op('fit');
		$format = $this->_val_format($format);
		return pack('C', $op) .
			pack('n', $width) .
			pack('n', $height) .
			pack('C', $format) .
			$path;
	}

	private function _build_fill($path, $width, $height, $anchor, $format="jpeg_gif")
	{
		$op = $this->_val_op('fill');
		$format = $this->_val_format($format);
		$anchor = $this->_val_anchor($anchor);

		return pack('C', $op) .
			pack('n', $width) .
			pack('n', $height) .
			pack('C', $anchor) .
			pack('C', $format) .
			$path;
	}

	private function _base64url_encode($data)
	{
		return rtrim(strtr(base64_encode($data), '+/', '-_'), '=');
	}

	private function _base64url_decode($data)
	{
		return base64_decode(strtr($data, '-_', '+/'));
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

	private function _str_op($op)
	{
		switch ($op) {
		case 1: return 'ori';
		case 2: return 'resize';
		case 3: return 'crop';
		case 4: return 'fit';
		case 5: return 'fill';
		default: return '';
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
		case 'jpeg_gif': return 4;
		default: throw new \Exception("undefined format: {$format}");
		}
	}

	private function _str_format($format)
	{
		switch ($format) {
		case 0: return '';
		case 1: return 'jpg';
		case 2: return 'png';
		case 3: return 'gif';
		case 4: return 'jpeg_gif';
		default: return '';
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

	private function _str_anchor($anchor)
	{
		switch ($anchor) {
		case 1: return 'topleft';
		case 2: return 'top';
		case 3: return 'topright';
		case 4: return 'left';
		case 5: return 'center';
		case 6: return 'right';
		case 7: return 'bottomleft';
		case 8: return 'bottom';
		case 9: return 'bottomright';
		default: return '';
		}
	}
}
