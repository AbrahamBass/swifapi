package dependencies

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/AbrahamBass/swifapi/internal/types"

	"github.com/spf13/cast"
)

func formatHeaderKey(key string) string {
	// Dividir la clave en sus componentes usando varios separadores
	parts := strings.FieldsFunc(key, func(r rune) bool {
		return r == '-' || r == '/' || r == '.' || r == '_' || r == ' '
	})

	if len(parts) == 0 {
		return ""
	}

	// Convertir la primera parte a minúsculas
	result := strings.ToLower(parts[0])

	// Convertir el resto de partes a title case y añadirlas
	for i := 1; i < len(parts); i++ {
		// Asegurarse de que la parte no esté vacía
		if len(parts[i]) > 0 {
			// Convertir primera letra a mayúscula y el resto a minúscula
			part := strings.ToLower(parts[i])
			firstChar := strings.ToUpper(string(part[0]))
			restChars := ""
			if len(part) > 1 {
				restChars = part[1:]
			}
			result += firstChar + restChars
		}
	}

	return result
}

func convertToType(value interface{}, targetType reflect.Type) (interface{}, error) {
	switch targetType.Kind() {
	case reflect.Int:
		return cast.ToIntE(value)
	case reflect.Int64:
		return cast.ToInt64E(value)
	case reflect.Int8:
		return cast.ToInt8E(value)
	case reflect.Int16:
		return cast.ToInt16E(value)
	case reflect.Int32:
		return cast.ToInt32E(value)
	case reflect.Uint:
		return cast.ToUintE(value)
	case reflect.Uint8:
		return cast.ToUint8E(value)
	case reflect.Uint16:
		return cast.ToUint16E(value)
	case reflect.Uint32:
		return cast.ToUint32E(value)
	case reflect.Uint64:
		return cast.ToUint64E(value)
	case reflect.Float64:
		return cast.ToFloat64E(value)
	case reflect.Float32:
		return cast.ToFloat32E(value)
	case reflect.String:
		return cast.ToStringE(value)
	case reflect.Bool:
		return cast.ToBoolE(value)
	default:
		return nil, fmt.Errorf("Tipo no soportado: %s", targetType.Kind())
	}
}

func buildMiddlewareChain(
	finalHandler func(types.IMiddlewareContext),
	middlewares []types.Middleware,
) func(types.IMiddlewareContext) {
	chain := finalHandler
	for i := len(middlewares) - 1; i >= 0; i-- {
		mw := middlewares[i]
		next := chain
		chain = func(ctx types.IMiddlewareContext) {
			mw(ctx, func() {
				next(ctx)
			})
		}
	}
	return chain
}
