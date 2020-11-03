<?php
define("SUPPORTED_ACTIONS", array("auth", "publish", "republish"));
define("API_KEY", getenv("API_KEY"));
define("SIGNATURE_SECRET", getenv("SIGNATURE_SECRET"));

header("Content-Type: application/json");

function json_response($code, $data)
{
	http_response_code($code);
	echo json_encode($data);
	exit(0);
}

function json_error_response($code, $message)
{
	json_response($code, array("error" => $message));
}

if($_SERVER["REQUEST_METHOD"] != "POST")
	json_error_response(400, "bad request");

$headers = apache_request_headers();

$platform = isset($headers["X-Kapost-Platform"]) ? $headers["X-Kapost-Platform"] : "";
$action = isset($headers["X-Kapost-Action"]) ? $headers["X-Kapost-Action"] : "";
$signature = isset($headers["X-Kapost-Signature"]) ? $headers["X-Kapost-Signature"] : "";
$api_key = preg_replace("/^Bearer\s+/", "", $headers["Authorization"]);
$body = file_get_contents("php://input");

$expected_signature = implode(
	"=",
	array(
		"sha256",
		hash_hmac(
			"sha256",
			$platform . $action . $body,
			SIGNATURE_SECRET
		)
	)
);

if(empty($signature) || !hash_equals($signature, $expected_signature))
	json_error_response(403, "Invalid signature");

if(empty($api_key) || empty(API_KEY) || !hash_equals($api_key, API_KEY))
	json_error_response(401, "Invalid API Key");

if(!in_array($action, SUPPORTED_ACTIONS))
	json_error_response(405, "Action '$action' is not supported");

$payload = json_decode($body, true);

switch($action)
{
	case "auth":
	{
		# NOTE: the api_key could also be validated against some external service on
		# authentication in the app center and an "error" could be returned if it
		# was incorrect, etc.
		json_response(
			200,
			array(
				"capabilities" => array(
					"html" => true,
					"any_file" => true
				)
			)
		);
	}
	break;

	case "publish":
	{
		json_response(
			200,
			array(
				"metadata" => array(
					"external_id" => "abc33",
					"published_url" => "https://localhost/php-test"
				)
			)
		);
	}
	break;

	case "republish":
	{
		$external_id = $payload["action"]["external_id"] ?? null;
		if(!hash_equals($external_id, "abc33"))
			json_error_response(404, "Cannot republish because external id could not be found");

		json_response(
			200,
			array(
				"metadata" => array(
					"external_id" => "abc33",
					"published_url" => "https://localhost/php-test-republish"
				)
			)
		);
	}
	break;
}
?>
