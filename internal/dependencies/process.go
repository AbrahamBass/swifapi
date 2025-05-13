package dependencies

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/AbrahamBass/swiftapi/internal/responses"
	"github.com/AbrahamBass/swiftapi/internal/types"
	"github.com/AbrahamBass/swiftapi/internal/ws"

	"go.uber.org/zap"
)

var (
	paramCache  sync.Map
	dependCache sync.Map
)

type cacheKey struct {
	fnPtr uintptr
	fnSig string
}

func generateCacheKey(fn interface{}) cacheKey {
	v := reflect.ValueOf(fn)
	return cacheKey{
		fnPtr: v.Pointer(),
		fnSig: runtime.FuncForPC(v.Pointer()).Name(),
	}
}

func analyzeFunctionWithCache(fn interface{}) ([]param, error) {
	key := generateCacheKey(fn)

	if cached, exists := paramCache.Load(key); exists {
		return cached.([]param), nil
	}

	params, err := analyzeFunction(fn)
	if err != nil {
		return nil, err
	}

	paramCache.Store(key, params)
	return params, nil
}

func analyzeDependenciesWithCache(fn interface{}) (*dependant, error) {
	key := generateCacheKey(fn)

	if cached, exists := dependCache.Load(key); exists {
		return cached.(*dependant), nil
	}

	params, err := analyzeFunctionWithCache(fn)
	if err != nil {
		return nil, err
	}

	depend, err := getDependant(params)
	if err != nil {
		return nil, err
	}

	dependCache.Store(key, depend)
	return depend, nil
}

func analyzeFunction(fn interface{}) ([]param, error) {
	fnType := reflect.TypeOf(fn)
	if fnType.Kind() != reflect.Func {
		return nil, fmt.Errorf("se requiere una función")
	}

	funcStr, err := getFunctionSource(fn)
	if err != nil {
		return nil, err
	}

	return extractParamsWithTypes(funcStr, fnType)
}

func getFunctionSource(fn interface{}) (string, error) {
	pc := runtime.FuncForPC(reflect.ValueOf(fn).Pointer())
	file, line := pc.FileLine(pc.Entry())

	fset := token.NewFileSet()
	parsedFile, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
	if err != nil {
		return "", err
	}

	var funcDecl *ast.FuncDecl
	ast.Inspect(parsedFile, func(n ast.Node) bool {
		if fd, ok := n.(*ast.FuncDecl); ok && fset.Position(fd.Pos()).Line <= line && fset.Position(fd.End()).Line >= line {
			funcDecl = fd
			return false
		}
		return true
	})

	if funcDecl == nil {
		return "", fmt.Errorf("función no encontrada en el AST")
	}

	var buf bytes.Buffer
	if err := printer.Fprint(&buf, fset, funcDecl); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func extractParamsWithTypes(funcStr string, fnType reflect.Type) ([]param, error) {
	wrapped := "package main\n\n" + funcStr
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", wrapped, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var funcDecl *ast.FuncDecl
	for _, decl := range file.Decls {
		if fd, ok := decl.(*ast.FuncDecl); ok {
			funcDecl = fd
			break
		}
	}

	if funcDecl == nil {
		return nil, fmt.Errorf("declaración de función no encontrada")
	}

	params := make([]param, 0)
	if funcDecl.Type.Params != nil {
		paramIndex := 0
		for _, field := range funcDecl.Type.Params.List {

			for _, name := range field.Names {
				if paramIndex >= fnType.NumIn() {
					return nil, fmt.Errorf("discrepancia en número de parámetros")
				}

				params = append(params, param{
					I:           paramIndex,
					Name:        name.Name,
					ReflectType: fnType.In(paramIndex),
				})
				paramIndex++
			}
		}
	}

	return params, nil
}

func generic(t reflect.Type) (string, reflect.Type, error) {
	if t.Kind() != reflect.Struct {
		return "", nil, fmt.Errorf("el tipo %s no es una estructura", t)
	}

	baseName := ""

	splitType := strings.Split(t.Name(), "[")
	if len(splitType) != 0 {
		baseName = splitType[0]
	}

	field, ok := t.FieldByName("Value")
	if !ok {
		return "", nil, fmt.Errorf("el tipo %s no tiene campo 'Value'", baseName)
	}

	if field.Type == nil {
		return "", nil, fmt.Errorf("tipo genérico inválido en campo Value")
	}

	return baseName, field.Type, nil
}

func getDependant(params []param) (*dependant, error) {
	depends := newDependant()

	for _, p := range params {
		baseName, fieldType, err := generic(p.ReflectType)

		if err != nil {
			return nil, err
		}

		model := &model{
			I:           p.I,
			Name:        p.Name,
			Type:        fieldType,
			ReflectType: p.ReflectType,
		}
		tag := types.TagType(baseName)
		if !addNonFieldParamToDependency(
			model,
			depends,
		) {
			addParamToFields(
				tag,
				model,
				depends,
			)
		}

	}

	return depends, nil
}

func addNonFieldParamToDependency(
	field *model,
	dependant *dependant,
) bool {
	if field.Type == reflect.TypeOf(&http.Request{}) {
		dependant.Request = field
		return true
	} else if field.Type == reflect.TypeOf((*ws.WebsocketManager)(nil)) {
		dependant.Websocket = field
		return true
	}
	return false
}

func addParamToFields(
	tag types.TagType,
	field *model,
	dependant *dependant,
) {
	switch tag {
	case types.TagQuery:
		dependant.QueryParams = append(dependant.QueryParams, field)
	case types.TagPath:
		dependant.PathParams = append(dependant.PathParams, field)
	case types.TagBody:

		if len(dependant.BodyParams) > 0 {
			panic("only one body parameter is allowed")
		}

		if len(dependant.FormParams) > 0 {
			panic("cannot mix body and form parameters")
		}

		dependant.BodyParams = append(dependant.BodyParams, field)
	case types.TagCookie:
		dependant.CookieParams = append(dependant.CookieParams, field)
	case types.TagHeader:
		dependant.HeaderParams = append(dependant.HeaderParams, field)
	case types.TagForm:

		if len(dependant.BodyParams) > 0 {
			panic("cannot add form parameters when body is already present")
		}

		dependant.FormParams = append(dependant.FormParams, field)
	case types.TagService:
		dependant.ServiceParams = append(dependant.ServiceParams, field)
	case types.TagContext:
		dependant.ContextParams = append(dependant.ContextParams, field)
	}
}

func safeSetField(field *model, value interface{}) error {

	genericType := field.ReflectType.Field(0).Type

	instancePtr := reflect.New(field.ReflectType)
	instance := instancePtr.Elem()

	val := reflect.ValueOf(value)

	if field.Type.Kind() == reflect.Ptr {
		if val.Kind() != reflect.Ptr {
			if val.Type().AssignableTo(field.Type.Elem()) {
				ptr := reflect.New(val.Type())
				ptr.Elem().Set(val)
				val = ptr
			} else {
				return fmt.Errorf("type mismatch: cannot convert %s to %s", val.Type(), field.Type)
			}
		}
	} else {
		if val.Kind() == reflect.Ptr {
			if val.IsNil() {
				return fmt.Errorf("nil pointer for non-pointer type %s", field.Type)
			}
			if val.Type().Elem().AssignableTo(field.Type) {
				val = val.Elem()
			} else {
				return fmt.Errorf("type mismatch: cannot dereference %s to %s", val.Type(), field.Type)
			}
		}
	}

	if !val.Type().AssignableTo(genericType) {
		return fmt.Errorf("type conversion error: cannot assign %T to %s", value, instance.Type())
	}

	instance.Field(0).Set(val)
	field.Default = instance.Interface()

	return nil
}

func extractBody(r *http.Request) (interface{}, []*issue) {
	var issues []*issue
	var body interface{}
	loc := []string{"body"}

	if r.Body == nil || r.Body == http.NoBody {
		return nil, issues
	}

	contentType := r.Header.Get("Content-Type")

	rawBody, err := io.ReadAll(r.Body)
	if err != nil {
		issues = append(issues, newIssue(
			loc,
			"Failed to read request body",
			types.BodyRead,
		))
		return nil, issues
	}

	r.Body = io.NopCloser(bytes.NewBuffer(rawBody))

	if len(rawBody) == 0 {
		return nil, issues
	}

	switch {
	case strings.Contains(contentType, "multipart/form-data"):
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			issues = append(issues, newIssue(
				loc,
				fmt.Sprintf("Failed to parse multipart form: %s", err.Error()),
				types.Multipart,
			))
		} else {
			body = r.MultipartForm
		}
	case contentType == "application/x-www-form-urlencoded":
		if err := r.ParseForm(); err != nil {
			issues = append(issues, newIssue(
				loc,
				fmt.Sprintf("Failed to parse form data: %s", err.Error()),
				types.URLEncoded,
			))
		} else {
			body = r.PostForm
		}
	default:
		body = rawBody
	}

	return body, issues
}

func processInputFields(fields []*model, input map[string]string, location string) []*issue {
	var issues []*issue

	for _, field := range fields {
		rawValue, exists := input[field.Name]
		loc := []string{location, field.Name}

		if !exists {
			issues = append(issues, newIssue(
				loc,
				"field required",
				types.Missing,
			))
			continue
		}

		convertedValue, err := convertToType(rawValue, field.Type)
		if err != nil {
			issues = append(issues, newIssue(
				loc,
				fmt.Sprintf("Conversion error: %v", err),
				types.General,
			))
			continue
		}

		if err := safeSetField(field, convertedValue); err != nil {
			issues = append(issues, newIssue(loc, err.Error(), types.TypeError))
		}
	}

	return issues
}

func parseParams(req *http.Request) map[string]string {
	params := make(map[string]string)
	if rv := req.Context().Value(1); rv != nil {
		mapParams, ok := rv.(map[string]string)
		if ok {
			for k, v := range mapParams {
				params[k] = v
			}
		}
	}
	return params
}

func parseQuery(req *http.Request) map[string]string {
	queryParams := req.URL.Query()
	queryMap := make(map[string]string)
	for key := range queryParams {
		queryMap[key] = queryParams.Get(key)
	}
	return queryMap
}

func parseHeaders(r *http.Request) map[string]string {
	headers := make(map[string]string)
	for key, values := range r.Header {
		if len(values) > 0 {
			camelCaseKey := formatHeaderKey(key)
			headers[camelCaseKey] = strings.Join(values, ", ")
		}
	}
	return headers
}

func parseCookies(r *http.Request) map[string]string {
	cookies := make(map[string]string)
	for _, cookie := range r.Cookies() {
		cookies[cookie.Name] = cookie.Value
	}
	return cookies
}

func handleUploadType(field *model, b map[string][]*multipart.FileHeader, loc []string) *issue {
	files, ok := b[field.Name]
	if !ok || len(files) == 0 {
		return newIssue(loc, field.Name, types.Missing)
	}

	if err := safeSetField(field, files[0]); err != nil {
		return newIssue(loc, err.Error(), types.TypeError)
	}

	return nil
}

func handleMultipartField(field *model, b map[string][]string, loc []string) *issue {
	values, ok := b[field.Name]
	if !ok || len(values) == 0 {
		return newIssue(loc, field.Name, types.Missing)
	}

	convertedValue, err := convertToType(values[0], field.Type)
	if err != nil {
		return newIssue(loc, err.Error(), types.General)
	}

	if err := safeSetField(field, convertedValue); err != nil {
		return newIssue(loc, err.Error(), types.TypeError)
	}

	return nil
}

func parseMultipart(field *model, body interface{}) *issue {
	loc := []string{"form", field.Name}

	if body == nil {
		return newIssue(
			loc,
			"form data is required",
			types.Missing,
		)
	}

	switch b := body.(type) {
	case *multipart.Form:
		if field.Type == reflect.TypeOf((*types.UploadFile)(nil)).Elem() {
			return handleUploadType(field, b.File, loc)
		} else {
			return handleMultipartField(field, b.Value, loc)
		}
	case url.Values:
		return handleMultipartField(field, b, loc)
	default:
		return newIssue(
			loc,
			fmt.Sprintf("Unsupported form type: %T", body),
			types.UnsupportedType,
		)
	}
}

func processMultipart(modelField []*model, body interface{}) []*issue {
	var issues []*issue
	for _, field := range modelField {
		if issue := parseMultipart(field, body); issue != nil {
			issues = append(issues, issue)
		}
	}
	return issues
}

func parseBody(field *model, body interface{}) *issue {
	loc := []string{"body", field.Name}

	if body == nil {
		return newIssue(
			loc,
			"request body is required",
			types.Missing,
		)
	}

	b, ok := body.([]byte)
	if !ok {
		return newIssue(
			loc,
			fmt.Sprintf("invalid body type: %T", body),
			types.InvalidType,
		)
	}

	if len(b) == 0 {
		return newIssue(
			loc,
			"body cannot be empty",
			types.Empty,
		)
	}

	switch field.Type.Kind() {
	case reflect.Slice:
		return handleSliceType(field, b, loc)
	case reflect.String:
		return handleStringType(field, b, loc)
	case reflect.Struct, reflect.Ptr:
		return handleStructType(field, b, loc)
	default:
		return newIssue(
			loc,
			fmt.Sprintf("unsupported body type: %s", field.Type.Kind()),
			types.Unsupported,
		)
	}
}

func handleSliceType(field *model, b []byte, loc []string) *issue {
	if err := safeSetField(field, b); err != nil {
		return newIssue(loc, err.Error(), types.TypeError)
	}
	return nil
}

func handleStringType(field *model, b []byte, loc []string) *issue {
	str := string(b)
	if err := safeSetField(field, str); err != nil {
		return newIssue(loc, err.Error(), types.TypeError)
	}
	return nil
}

func handleStructType(field *model, b []byte, loc []string) *issue {
	targetType := field.Type
	if targetType.Kind() == reflect.Ptr {
		targetType = targetType.Elem()
	}

	decoded := reflect.New(targetType).Interface()
	if err := json.Unmarshal(b, decoded); err != nil {
		return parseJSONError(err, loc)
	}

	if field.Type.Kind() != reflect.Ptr {
		decoded = reflect.ValueOf(decoded).Elem().Interface()
	}

	if err := safeSetField(field, decoded); err != nil {
		return newIssue(loc, err.Error(), types.TypeError)
	}

	return nil
}

func parseJSONError(err error, loc []string) *issue {
	var syntaxErr *json.SyntaxError
	var typeErr *json.UnmarshalTypeError
	var unmarshalErr *json.InvalidUnmarshalError

	switch {
	case errors.As(err, &syntaxErr):
		return newIssue(
			loc,
			fmt.Sprintf("Invalid JSON syntax at position %d: %v", syntaxErr.Offset, syntaxErr.Error()),
			types.Syntax,
		)
	case errors.As(err, &typeErr):
		return newIssue(
			loc,
			fmt.Sprintf("Type mismatch at %s (expected %s, got %s)",
				typeErr.Field,
				typeErr.Type,
				typeErr.Value,
			),
			types.JSONType,
		)
	case errors.As(err, &unmarshalErr):
		return newIssue(
			loc,
			"Invalid unmarshal target type",
			types.Target,
		)
	default:
		return newIssue(
			loc,
			fmt.Sprintf("JSON decoding error: %v", err),
			types.Target,
		)
	}
}

func processBody(modelFiled []*model, body interface{}) []*issue {
	var issues []*issue
	for _, field := range modelFiled {
		if issue := parseBody(field, body); issue != nil {
			issues = append(issues, issue)
		}
	}
	return issues
}

func parseService(dig types.IDigContainer, logger *zap.Logger, field *model) *issue {
	loc := []string{"service", field.Name}
	instance, err := resolveDependency(dig, field.Type)
	if err != nil {
		logger.Fatal("🚨 Failed to resolve dependency",
			zap.Error(err),
		)
	}

	if err := safeSetField(field, instance); err != nil {
		return newIssue(loc, err.Error(), types.TypeError)
	}

	return nil
}

func resolveDependency(container types.IDigContainer, targetType reflect.Type) (interface{}, error) {
	var result interface{}

	fnType := reflect.FuncOf(
		[]reflect.Type{targetType},
		[]reflect.Type{},
		false,
	)

	fnValue := reflect.MakeFunc(fnType, func(args []reflect.Value) []reflect.Value {
		result = args[0].Interface()
		return nil
	})

	if err := container.Invoke(fnValue.Interface()); err != nil {
		return nil, err
	}

	return result, nil
}

func processService(dig types.IDigContainer, logger *zap.Logger, modelField []*model) []*issue {
	var issues []*issue
	for _, field := range modelField {
		if issue := parseService(dig, logger, field); issue != nil {
			issues = append(issues, issue)
		}
	}
	return issues
}

func parseContext(field *model, req *http.Request) *issue {
	loc := []string{"context", field.Name}
	value := req.Context().Value(field.Name)
	if value != nil {
		if err := safeSetField(field, value); err != nil {
			return newIssue(loc, err.Error(), types.TypeError)
		}
		return nil
	}
	return newIssue(
		loc,
		fmt.Sprintf("missing required context parameter: %s", field.Name),
		types.Missing,
	)
}

func processContext(modelField []*model, req *http.Request) []*issue {
	var issues []*issue
	for _, field := range modelField {
		if err := parseContext(field, req); err != nil {
			issues = append(issues, err)
		}
	}
	return issues
}

func solveDependant(
	webscoketManeger *ws.WebsocketManager,
	dig types.IDigContainer,
	logger *zap.Logger,
	req *http.Request,
	w *responses.ResponseWriter,
	dependant *dependant,
	body interface{},
) []*issue {
	var issues []*issue

	paramProcessors := []struct {
		params    []*model
		extractor func(*http.Request) map[string]string
		location  string
	}{
		{dependant.PathParams, parseParams, types.ParamLocationPath},
		{dependant.QueryParams, parseQuery, types.ParamLocationQuery},
		{dependant.HeaderParams, parseHeaders, types.ParamLocationHeader},
		{dependant.CookieParams, parseCookies, types.ParamLocationCookie},
	}

	for _, processor := range paramProcessors {
		if processor.params == nil {
			continue
		}

		values := processor.extractor(req)
		paramIssues := processInputFields(
			processor.params,
			values,
			processor.location,
		)
		issues = append(issues, paramIssues...)
	}

	if dependant.FormParams != nil {
		issues = append(
			issues,
			processMultipart(
				dependant.FormParams,
				body,
			)...,
		)
	}

	if dependant.BodyParams != nil {
		issues = append(
			issues,
			processBody(
				dependant.BodyParams,
				body,
			)...,
		)
	}

	if dependant.ServiceParams != nil {
		issues = append(
			issues,
			processService(
				dig,
				logger,
				dependant.ServiceParams,
			)...,
		)
	}

	if dependant.ContextParams != nil {
		issues = append(
			issues,
			processContext(
				dependant.ContextParams,
				req,
			)...,
		)
	}

	if dependant.Request != nil {
		if err := safeSetField(dependant.Request, req); err != nil {
			issues = append(
				issues,
				newIssue(
					[]string{"request", dependant.Request.Name},
					err.Error(),
					types.TypeError,
				),
			)
		}
	} else if dependant.Response != nil {
		if err := safeSetField(dependant.Response, w.W); err != nil {
			issues = append(
				issues,
				newIssue(
					[]string{"response", dependant.Response.Name},
					err.Error(),
					types.TypeError,
				),
			)
		}
	} else if dependant.Websocket != nil {
		if err := safeSetField(dependant.Websocket, webscoketManeger); err != nil {
			issues = append(
				issues,
				newIssue(
					[]string{"Websocket", dependant.Websocket.Name},
					err.Error(),
					types.TypeError,
				),
			)
		}
	}

	return issues
}

func processDependant(dependant *dependant) []reflect.Value {
	var allModels []*model

	sections := [][]*model{
		dependant.PathParams,
		dependant.QueryParams,
		dependant.BodyParams,
		dependant.CookieParams,
		dependant.HeaderParams,
		dependant.FormParams,
		dependant.ServiceParams,
		dependant.ContextParams,
	}

	for _, section := range sections {
		allModels = append(allModels, section...)
	}

	individualModels := []*model{
		dependant.Request,
		dependant.Response,
		dependant.Websocket,
	}

	for _, model := range individualModels {
		if model != nil {
			allModels = append(allModels, model)
		}
	}

	sort.Slice(allModels, func(i, j int) bool {
		return allModels[i].I < allModels[j].I
	})

	result := make([]reflect.Value, 0, len(allModels))
	for _, model := range allModels {
		var value reflect.Value

		if model.Default != nil {
			value = reflect.ValueOf(model.Default)
		} else {
			value = reflect.Zero(model.ReflectType)
		}

		result = append(result, value)
	}

	return result
}
