package main

import (
	"bytes"
	"fmt"
	"go/format"
	"log"
	"os"
	"reflect"
	"strings"

	"github.com/FreekingDean/proxmox-api-go/proxmox/access/groups"
	"github.com/FreekingDean/proxmox-api-go/proxmox/access/roles"
	"github.com/FreekingDean/proxmox-api-go/proxmox/access/users"
	"github.com/FreekingDean/proxmox-api-go/proxmox/cluster/firewall"
	"github.com/FreekingDean/proxmox-api-go/proxmox/cluster/firewall/aliases"
	fgroups "github.com/FreekingDean/proxmox-api-go/proxmox/cluster/firewall/groups"
	frules "github.com/FreekingDean/proxmox-api-go/proxmox/cluster/firewall/rules"
)

var (
	//d  = domains.New(nil)
	g  = groups.New(nil)
	r  = roles.New(nil)
	cf = firewall.New(nil)
	a  = aliases.New(nil)
	u  = users.New(nil)
	fg = fgroups.New(nil)
	ru = frules.New(nil)
)

type Source struct {
	function any
	funcName string
}

var datasources = map[string]map[string]Source{
	"access": {
		//"domains": d.Index,
		//"domain": d.Find,
		"groups": {g.Index, "Index"},
		"group":  {g.Find, "Find"},
		"roles":  {r.Index, "Index"},
		"role":   {r.Find, "Find"},
		"users":  {u.Index, "Index"},
		"user":   {u.Find, "Find"},
	},
	"cluster": {
		"firewalls": {cf.Index, "Index"},
		//"firewall":  {cf.Find, "Find"},
	},
	"cluster/firewall": {
		"aliases": {a.Index, "Index"},
		"alias":   {a.Find, "Find"},
		"groups":  {fg.Index, "Index"},
		"group":   {fg.Find, "Find"},
		"rules":   {ru.Index, "Index"},
		"rule":    {ru.Find, "Find"},
	},
}

func main() {
	for domain, sources := range datasources {
		for name, source := range sources {
			genData(fmt.Sprintf("%s/%s", domain, name), source.funcName, source.function)
		}
	}
}

func genData(name string, funcName string, function interface{}) {
	packageName := snake(name)
	filename := fmt.Sprintf("data_%s.go", packageName)
	f, err := os.OpenFile("proxmox/"+filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	buf := &bytes.Buffer{}
	p := func(msg string, v ...any) {
		fmt.Fprintf(buf, msg+"\n", v...)
	}

	t := reflect.TypeOf(function)
	if t.Kind() != reflect.Func {
		log.Fatal("Function not actual function")
	}
	pkgPath := ""
	var inType, outType reflect.Type
	if t.NumIn() > 1 {
		tt := t.In(1)
		inType = tt.Elem()
	}
	tt := t.Out(0)
	outType = tt.Elem()
	pkgPath = outType.PkgPath()
	p("package proxmox")
	p("")
	p("import (")
	p("  \"context\"")
	p("")
	p("  \"github.com/hashicorp/terraform-plugin-framework/datasource\"")
	p("  \"github.com/hashicorp/terraform-plugin-framework/datasource/schema\"")
	p("  \"github.com/hashicorp/terraform-plugin-framework/types\"")
	p("")
	p("  \"github.com/FreekingDean/proxmox-api-go/proxmox\"")
	p("  \"%s\"", pkgPath)
	p(")")
	p("")
	p("var (")
	p("  _ datasource.DataSource = &%sDataSource{}", camel(name))
	p(")")
	p("")
	p("func init() {")
	p("  datasources = append(datasources, New%sDataSource)", upFirst(camel(name)))
	p("}")
	p("")
	p("type %sDataSource struct {", camel(name))
	parts := strings.Split(pkgPath, "/")
	pkgName := parts[len(parts)-1]
	p("  client *%s.Client", pkgName)
	p("}")
	p("")
	p("func New%sDataSource() datasource.DataSource {", upFirst(camel(name)))
	p("  return &%sDataSource{}", camel(name))
	p("}")
	p("")
	p("type %sModel struct {", camel(name))
	e := &extraTypes{
		types: make(map[string]reflect.Type),
	}
	if inType != nil {
		for i := 0; i < inType.NumField(); i++ {
			field := inType.Field(i)
			p("  %s %s `tfsdk:\"%s\"`", field.Name, e.typeStr("types.", field.Name, field.Type), snake(field.Name))
		}
	}
	if outType.Kind() == reflect.Slice {
		p(" %s []%s `tfsdk:\"%s\"`", upFirst(camel(name)), e.typeStr("types.", upFirst(camel(name)), outType.Elem()), snake(name))
	} else if outType.Kind() == reflect.Struct {
		for i := 0; i < outType.NumField(); i++ {
			field := outType.Field(i)
			p(" %s %s `tfsdk:\"%s\"`", field.Name, e.typeStr("types.", field.Name, field.Type), snake(field.Name))
		}
	} else if outType.Kind() == reflect.Map {
		p(" Attrs map[string]types.String `tfsdk:\"attrs\"`")
	}
	p("}")
	p("")
	for len(e.types) > 0 {
		for k, t := range e.types {
			p("type %s struct {", k)
			for i := 0; i < t.NumField(); i++ {
				field := t.Field(i)
				p(" %s %s `tfsdk:\"%s\"`", field.Name, e.typeStr("types.", field.Name, field.Type), snake(field.Name))
			}
			p("}")
			p("")
			delete(e.types, k)
		}
	}
	p("func (d *%sDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {", camel(name))
	p("  if client, ok := req.ProviderData.(*proxmox.Client); ok {")
	p("    d.client = %s.New(client)", pkgName)
	p("  }")
	p("}")
	p("")
	p("func (d *%sDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {", camel(name))
	p("  resp.TypeName = req.ProviderTypeName + \"_%s\"", snake(name))
	p("}")
	p("")
	p("func (d *%sDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {", camel(name))
	p("  resp.Schema = schema.Schema{")
	p("    Attributes: map[string]schema.Attribute{")
	if inType != nil {
		for i := 0; i < inType.NumField(); i++ {
			field := inType.Field(i)
			p("\"%s\": %s", snake(field.Name), schemaStr(field.Type, false, false))
		}
	}
	if outType.Kind() == reflect.Ptr {
		outType = outType.Elem()
	}
	if outType.Kind() == reflect.Slice {
		if outType.Elem().Kind() == reflect.Struct || outType.Elem().Kind() == reflect.Ptr && outType.Elem().Elem().Kind() == reflect.Struct {
			p("\"%s\": schema.ListNestedAttribute{", snake(name))
			p("  Computed: true,")
			p("  NestedObject: schema.NestedAttributeObject{")
			p("    Attributes: map[string]schema.Attribute{")
			tt := outType.Elem().Elem()
			for i := 0; i < tt.NumField(); i++ {
				field := tt.Field(i)
				p("\"%s\": %s", snake(field.Name), schemaStr(field.Type, true, false))
			}
			p("},\n")
			p("},\n")
			p("},\n")
		} else {
			p("\"%s\": schema.ListAttribute{", snake(name))
			p("  Computed: true,")
			el := outType.Elem()
			if el.Kind() == reflect.Ptr {
				el = el.Elem()
			}
			if el.Kind() == reflect.Map {
				p("  ElementType: types.MapType{")
				p("    ElemType: %sType,", e.typeStr("types.", "types.", el.Elem()))
				p("},")
			} else {
				p("  ElementType: %s", e.typeStr("types.", "types.", el.Elem()))
			}
			p("},")
		}
	} else if outType.Kind() == reflect.Struct {
		for i := 0; i < outType.NumField(); i++ {
			field := outType.Field(i)
			p("\"%s\": %s", snake(field.Name), schemaStr(field.Type, true, false))
		}
	} else if outType.Kind() == reflect.Map {
		p("\"attrs\": schema.MapAttribute{\nElementType: types.StringType,\nComputed: true,\n},")
	}
	p("    },")
	p("  }")
	p("}")
	p("")
	p("func (d *%sDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {", camel(name))
	p("  var state %sModel", camel(name))
	p("")
	if inType != nil {
		p("diags := req.Config.Get(ctx, &state)")
		p("resp.Diagnostics.Append(diags...)")
		printDiagCheck(p)
	}
	p("%s, err := d.client.%s(", camel(name), funcName)
	p("ctx,")
	if inType != nil {
		p("&%s.%s{", pkgName, inType.Name())
		for i := 0; i < inType.NumField(); i++ {
			field := inType.Field(i)
			p("%s: state.%s.%s(),", field.Name, field.Name, e.typeStr("Value", field.Name, field.Type))
		}
		p("  ")
		p("},")
	}
	p(")")
	p("")
	p("if err != nil {")
	p("	resp.Diagnostics.AddError(")
	p("		\"Unable to Read Proxmox %s\",", upFirst(camel(name)))
	p("		err.Error(),")
	p("	)")
	p("	return")
	p("}")
	p("")

	utilNeeded := printSave(p, outType, name)
	if utilNeeded {
		all := buf.String()
		lines := strings.Split(all, "\n")
		lines = append(lines[0:12], lines[11:len(lines)-1]...)
		lines[11] = "\"github.com/FreekingDean/terraform-provider-proxmox/internal/utils\""
		buf = bytes.NewBufferString(strings.Join(lines, "\n"))
	}
	//p("&%s.FindRequest{", camel(name), pkgName)
	//p("	Name: state.Name.ValueString(),")
	//p("})")
	p("")
	if inType != nil {
		p("diags = resp.State.Set(ctx, &state)")
	} else {
		p("diags := resp.State.Set(ctx, &state)")
	}
	p("resp.Diagnostics.Append(diags...)")
	p("if resp.Diagnostics.HasError() {")
	p("	return")
	p("}")
	p("}")
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(f, "%s", formatted)
}

func printSave(p printer, t reflect.Type, name string) bool {
	out := false
	e := &extraTypes{
		types: make(map[string]reflect.Type),
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() == reflect.Slice {
		//p(" %s []%s `tfsdk:\"%s\"`", upFirst(camel(name)), e.typeStr("types.", upFirst(camel(name)), t.Elem()), snake(name))
		p("for _, e := range *%s {", camel(name))
		elem := t.Elem()
		if elem.Kind() == reflect.Ptr {
			elem = elem.Elem()
		}
		if elem.Kind() == reflect.Struct {
			if t.Elem().Kind() == reflect.Ptr {
				t = t.Elem()
			}
			p("eState := %s{}", e.typeStr("types.", upFirst(camel(name)), t.Elem()))
			for i := 0; i < t.Elem().NumField(); i++ {
				field := t.Elem().Field(i)
				if field.Type.Kind() == reflect.Ptr {
					p("if e.%s != nil {", field.Name)
					tStr := e.typeStr("types.", field.Name, field.Type)
					b := ""
					be := ""
					if tStr == "types.Bool" {
						b = "bool("
						be = ")"
					}
					p("eState.%s = %sValue(%s*e.%s%s)", field.Name, tStr, b, field.Name, be)
					p("}")
				} else {
					tStr := e.typeStr("types.", field.Name, field.Type)
					b := ""
					be := ""
					if tStr == "types.Bool" {
						b = "bool("
						be = ")"
					}
					p("eState.%s = %sValue(%se.%s%s)", field.Name, tStr, b, field.Name, be)
				}
			}
		} else if elem.Kind() == reflect.Map {
			out = true
			p("eState, diag := utils.NormalizeMap(*e)")
			p("resp.Diagnostics.Append(diag...)")
			printDiagCheck(p)
		} else {
			p("eState := %sValue(e)", e.typeStr("types.", "", t.Elem()))
		}
		p("state.%s = append(state.%s, eState)", upFirst(camel(name)), upFirst(camel(name)))
		p("}")
	} else if t.Kind() == reflect.Struct {
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if field.Type.Kind() == reflect.Slice {
				p("for _, e := range %s.%s {", camel(name), field.Name)
				p("eState := %sValue(e)", e.typeStr("types.", field.Name, field.Type.Elem()))
				p("state.%s = append(state.%s, eState)", field.Name, field.Name)
				p("}")
			} else {
				if field.Type.Kind() == reflect.Ptr {
					p("if %s.%s != nil {", camel(name), field.Name)
					tStr := e.typeStr("types.", field.Name, field.Type.Elem())
					b := ""
					be := ""
					if tStr == "types.Bool" {
						b = "bool("
						be = ")"
					}
					p("state.%s = %sValue(%s*%s.%s%s)", field.Name, tStr, b, camel(name), field.Name, be)
					p("}")
				} else {
					tStr := e.typeStr("types.", field.Name, field.Type)
					b := ""
					be := ""
					if tStr == "types.Bool" {
						b = "bool("
						be = ")"
					}
					p("state.%s = %sValue(%s%s.%s%s)", field.Name, tStr, b, camel(name), field.Name, be)
				}
			}
		}
	} else if t.Kind() == reflect.Map {
		out = true
		p("state.Attrs, diags = utils.NormalizeMap(*%s)", camel(name))
		p("resp.Diagnostics.Append(diags...)")
		printDiagCheck(p)
	}
	return out
}

func schemaStr(t reflect.Type, computed bool, optional bool) string {
	out := ""
	scope := "Required"
	if computed {
		scope = "Computed"
	} else if optional {
		scope = "Optional"
	}
	switch t.Kind() {
	case reflect.Slice:
		if t.Elem().Kind() == reflect.Struct || t.Elem().Kind() == reflect.Ptr {
			out = fmt.Sprintf("schema.ListNestedAttribute{\n  %s: true,\n", scope)
			out += fmt.Sprintf("  NestedObject: %s\n},", schemaStr(t.Elem(), computed, optional))
		} else {
			out = fmt.Sprintf("schema.ListAttribute{\n  %s: true,\n", scope)
			out += fmt.Sprintf("  ElementType: %sType,\n},", (&extraTypes{}).typeStr("types.", "", t.Elem()))
		}
	case reflect.Bool:
		out = fmt.Sprintf("schema.BoolAttribute{\n  %s: true,\n},", scope)
	case reflect.String:
		out = fmt.Sprintf("schema.StringAttribute{\n  %s: true,\n},", scope)
	case reflect.Struct:
		out = "schema.NestedAttributeObject{\n"
		out += "Attributes: map[string]schema.Attribute{\n"
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			out += fmt.Sprintf("\"%s\": %s\n", snake(field.Name), schemaStr(field.Type, computed, optional))
		}
		out += "},\n"
		out += "},"
	case reflect.Ptr:
		return schemaStr(t.Elem(), computed, optional)
	case reflect.Int:
		out = fmt.Sprintf("schema.Int64Attribute{\n  %s: true,\n},", scope)
	case reflect.Map:
		out = fmt.Sprintf("schema.MapAttribute{\n  %s: true,\nElementType: %s,\n},", scope, (&extraTypes{}).typeStr("types.", "", t.Elem()))
	default:
		log.Fatalf("Unkown schema type %s", t.Kind())
	}
	return out
}

type extraTypes struct {
	types map[string]reflect.Type
}

func (e *extraTypes) typeStr(prefix string, name string, k reflect.Type) string {
	switch k.Kind() {
	case reflect.Int:
		return fmt.Sprintf("%sInt64", prefix)
	case reflect.Bool:
		return fmt.Sprintf("%sBool", prefix)
	case reflect.String:
		return fmt.Sprintf("%sString", prefix)
	case reflect.Slice:
		return fmt.Sprintf("[]%s", e.typeStr(prefix, name, k.Elem()))
	case reflect.Ptr:
		return fmt.Sprintf("%s", e.typeStr(prefix, name, k.Elem()))
	case reflect.Interface:
		return fmt.Sprintf("%sString", prefix)
	case reflect.Map:
		return fmt.Sprintf("map[%s]%s",
			k.Key().String(),
			e.typeStr(prefix, name, k.Elem()))
	case reflect.Struct:
		e.types[name] = k
		return name
	}
	log.Fatalf("Unkown type type %s", k.Kind())
	return ""
}

func snake(i string) string {
	i = strings.ToLower(i)
	c := false
	i = strings.Replace(i, "/", "_", -1)
	i = strings.Replace(i, "-", "_", -1)
	buf := ""
	for _, ch := range i {
		if ch >= 'A' && ch <= 'Z' {
			if c {
				buf = buf + "_"
			}
			ch = ch - 'A' + 'a'
			c = true
		}
		if ch == '_' {
			c = false
		}
		buf = fmt.Sprintf("%s%c", buf, ch)
	}
	return buf
}

func camel(i string) string {
	i = snake(i)
	buf := ""
	c := false
	for _, ch := range i {
		if c {
			ch -= 'a'
			ch += 'A'
			c = false
		}
		if ch == '_' {
			c = true
			continue
		}
		buf = fmt.Sprintf("%s%c", buf, ch)
	}
	return buf
}

func upFirst(i string) string {
	return fmt.Sprintf("%c%s", i[0]-'a'+'A', i[1:])
}

type printer func(string, ...any)

func printDiagCheck(p printer) {
	p("if resp.Diagnostics.HasError() {")
	p("	return")
	p("}")
	p("")
}
