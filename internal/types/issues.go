package types

type IssueType string

const (
	Missing         IssueType = "missing"          // Falta un campo
	Invalid         IssueType = "invalid"          // Valor no válido
	TypeError       IssueType = "type_error"       // Error de tipo
	Unsupported     IssueType = "unsupported"      // No soportado
	Empty           IssueType = "empty"            // Campo vacío
	Multipart       IssueType = "multipart"        // Error en multipart
	URLEncoded      IssueType = "urlencoded"       // Error en urlencoded
	Syntax          IssueType = "syntax"           // Error de sintaxis
	JSONType        IssueType = "json_type"        // Error en tipo JSON
	Target          IssueType = "target"           // Error en objetivo
	General         IssueType = "general"          // Error genérico
	BodyRead        IssueType = "body_read"        // Error al leer el cuerpo
	InvalidType     IssueType = "invalid_type"     // Tipo no válido
	UnsupportedType IssueType = "unsupported_type" // Tipo no soportado
)
