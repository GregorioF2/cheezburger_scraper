package consts

var HttpStatusCode map[string]map[string]int = map[string]map[string]int{
	"ServerError": {
		"InternalServerError": 500,
		"ServiceUnavailable":  503,
	},
	"ClientError": {
		"BadRequest":          400,
		"Unauthorized":        401,
		"Forbidden":           403,
		"NotFound":            404,
		"Conflict":            409,
		"Gone":                410,
		"UnprocessableEntity": 422,
		"Locked":              423,
		"FailedDependency":    424,
	},
	"Success": {
		"Ok":        200,
		"Created":   201,
		"NoContent": 204,
	},
	"Redirection": {
		"NotModified": 304,
	},
}
