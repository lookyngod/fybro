// +build android ios mobile

package gl

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/draw"

	"github.com/fyne-io/mobile/exp/f32"
	"github.com/fyne-io/mobile/gl"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/internal/cache"
	"fyne.io/fyne/v2/theme"
)

// Buffer represents a GL buffer
type Buffer gl.Buffer

// Program represents a compiled GL program
type Program gl.Program

// Texture represents an uploaded GL texture
type Texture gl.Texture

// NoTexture is the zero value for a Texture
var NoTexture = Texture(gl.Texture{0})

var textureFilterToGL = []int{gl.LINEAR, gl.NEAREST}

func (p *glPainter) logError() {
	if fyne.CurrentApp().Settings().BuildType() != fyne.BuildDebug {
		return
	}
	err := p.glctx().GetError()
	logGLError(uint32(err))
}

func (p *glPainter) glctx() gl.Context {
	return p.context.Context().(gl.Context)
}

func (p *glPainter) newTexture(textureFilter canvas.ImageScale) Texture {
	var texture = p.glctx().CreateTexture()
	p.logError()

	if int(textureFilter) >= len(textureFilterToGL) {
		fyne.LogError(fmt.Sprintf("Invalid canvas.ImageScale value (%d), using canvas.ImageScaleSmooth as default value", textureFilter), nil)
		textureFilter = canvas.ImageScaleSmooth
	}

	p.glctx().ActiveTexture(gl.TEXTURE0)
	p.glctx().BindTexture(gl.TEXTURE_2D, texture)
	p.logError()
	p.glctx().TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, textureFilterToGL[textureFilter])
	p.glctx().TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, textureFilterToGL[textureFilter])
	p.glctx().TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	p.glctx().TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	p.logError()

	return Texture(texture)
}

func (p *glPainter) imgToTexture(img image.Image, textureFilter canvas.ImageScale) Texture {
	switch i := img.(type) {
	case *image.Uniform:
		texture := p.newTexture(textureFilter)
		r, g, b, a := i.RGBA()
		r8, g8, b8, a8 := uint8(r>>8), uint8(g>>8), uint8(b>>8), uint8(a>>8)
		data := []uint8{r8, g8, b8, a8}
		p.glctx().TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, 1, 1, gl.RGBA,
			gl.UNSIGNED_BYTE, data)
		p.logError()
		return texture
	case *image.RGBA:
		if len(i.Pix) == 0 { // image is empty
			return NoTexture
		}

		texture := p.newTexture(textureFilter)
		p.glctx().TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, i.Rect.Size().X, i.Rect.Size().Y,
			gl.RGBA, gl.UNSIGNED_BYTE, i.Pix)
		p.logError()
		return texture
	default:
		rgba := image.NewRGBA(image.Rect(0, 0, img.Bounds().Dx(), img.Bounds().Dy()))
		draw.Draw(rgba, rgba.Rect, img, image.ZP, draw.Over)
		return p.imgToTexture(rgba, textureFilter)
	}
}

func (p *glPainter) SetOutputSize(width, height int) {
	p.glctx().Viewport(0, 0, width, height)
	p.logError()
}

func (p *glPainter) freeTexture(obj fyne.CanvasObject) {
	texture, ok := cache.GetTexture(obj)
	if !ok {
		return
	}

	p.glctx().DeleteTexture(gl.Texture(texture))
	p.logError()
	cache.DeleteTexture(obj)
}

func (p *glPainter) compileShader(source string, shaderType gl.Enum) (gl.Shader, error) {
	shader := p.glctx().CreateShader(shaderType)

	p.glctx().ShaderSource(shader, source)
	p.logError()
	p.glctx().CompileShader(shader)
	p.logError()

	status := p.glctx().GetShaderi(shader, gl.COMPILE_STATUS)
	if status == gl.FALSE {
		info := p.glctx().GetShaderInfoLog(shader)

		return shader, fmt.Errorf("failed to compile %v: %v", source, info)
	}

	return shader, nil
}

const (
	vertexShaderSource = `
    #version 100
    attribute vec3 vert;
    attribute vec2 vertTexCoord;
    varying highp vec2 fragTexCoord;

    void main() {
        fragTexCoord = vertTexCoord;

        gl_Position = vec4(vert, 1);
    }`

	fragmentShaderSource = `
    #version 100
    uniform sampler2D tex;

    varying highp vec2 fragTexCoord;

    void main() {
        gl_FragColor = texture2D(tex, fragTexCoord);
    }`

	vertexLineShaderSource = `
    #version 100
    attribute vec2 vert;
    attribute vec2 normal;
    
    uniform float lineWidth;

    varying highp vec2 delta;

    void main() {
        delta = normal * lineWidth;

        gl_Position = vec4(vert + delta, 0, 1);
    }`

	fragmentLineShaderSource = `
    #version 100
    uniform highp vec4 color;
    uniform highp float lineWidth;
    uniform highp float feather;

    varying highp vec2 delta;

    void main() {
        highp float alpha = color.a;
        highp float distance = length(delta);

        if (feather == 0.0 || distance <= lineWidth - feather) {
           gl_FragColor = color;
        } else {
           gl_FragColor = vec4(color.r, color.g, color.b, mix(color.a, 0.0, (distance - (lineWidth - feather)) / feather));
        }
    }`
)

func (p *glPainter) Init() {
	p.glctx().Disable(gl.DEPTH_TEST)
	p.glctx().Enable(gl.BLEND)

	vertexShader, err := p.compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		panic(err)
	}
	fragmentShader, err := p.compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		panic(err)
	}

	prog := p.glctx().CreateProgram()
	p.glctx().AttachShader(prog, vertexShader)
	p.glctx().AttachShader(prog, fragmentShader)
	p.glctx().LinkProgram(prog)
	p.logError()

	p.program = Program(prog)
	p.logError()

	vertexLineShader, err := p.compileShader(vertexLineShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		panic(err)
	}
	fragmentLineShader, err := p.compileShader(fragmentLineShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		panic(err)
	}

	lineProg := p.glctx().CreateProgram()
	p.glctx().AttachShader(lineProg, vertexLineShader)
	p.glctx().AttachShader(lineProg, fragmentLineShader)
	p.glctx().LinkProgram(lineProg)
	p.logError()

	p.lineProgram = Program(lineProg)
	p.logError()
}

func (p *glPainter) glClearBuffer() {
	r, g, b, a := theme.BackgroundColor().RGBA()
	max16bit := float32(255 * 255)
	p.glctx().ClearColor(float32(r)/max16bit, float32(g)/max16bit, float32(b)/max16bit, float32(a)/max16bit)
	p.glctx().Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	p.logError()
}

func (p *glPainter) glScissorOpen(x, y, w, h int32) {
	p.glctx().Scissor(x, y, w, h)
	p.glctx().Enable(gl.SCISSOR_TEST)
	p.logError()
}

func (p *glPainter) glScissorClose() {
	p.glctx().Disable(gl.SCISSOR_TEST)
	p.logError()
}

func (p *glPainter) glCreateBuffer(points []float32) Buffer {
	ctx := p.glctx()

	p.glctx().UseProgram(gl.Program(p.program))

	buf := ctx.CreateBuffer()
	p.logError()
	ctx.BindBuffer(gl.ARRAY_BUFFER, buf)
	p.logError()
	ctx.BufferData(gl.ARRAY_BUFFER, f32.Bytes(binary.LittleEndian, points...), gl.DYNAMIC_DRAW)
	p.logError()

	vertAttrib := ctx.GetAttribLocation(gl.Program(p.program), "vert")
	ctx.EnableVertexAttribArray(vertAttrib)
	ctx.VertexAttribPointer(vertAttrib, 3, gl.FLOAT, false, 5*4, 0)
	p.logError()

	texCoordAttrib := ctx.GetAttribLocation(gl.Program(p.program), "vertTexCoord")
	ctx.EnableVertexAttribArray(texCoordAttrib)
	ctx.VertexAttribPointer(texCoordAttrib, 2, gl.FLOAT, false, 5*4, 3*4)
	p.logError()

	return Buffer(buf)
}

func (p *glPainter) glCreateLineBuffer(points []float32) Buffer {
	ctx := p.glctx()

	p.glctx().UseProgram(gl.Program(p.lineProgram))

	buf := ctx.CreateBuffer()
	p.logError()
	ctx.BindBuffer(gl.ARRAY_BUFFER, buf)
	p.logError()
	ctx.BufferData(gl.ARRAY_BUFFER, f32.Bytes(binary.LittleEndian, points...), gl.DYNAMIC_DRAW)
	p.logError()

	vertAttrib := ctx.GetAttribLocation(gl.Program(p.lineProgram), "vert")
	ctx.EnableVertexAttribArray(vertAttrib)
	ctx.VertexAttribPointer(vertAttrib, 2, gl.FLOAT, false, 4*4, 0)
	p.logError()

	normalAttrib := ctx.GetAttribLocation(gl.Program(p.lineProgram), "normal")
	ctx.EnableVertexAttribArray(normalAttrib)
	ctx.VertexAttribPointer(normalAttrib, 2, gl.FLOAT, false, 4*4, 2*4)
	p.logError()

	return Buffer(buf)
}

func (p *glPainter) glFreeBuffer(b Buffer) {
	ctx := p.glctx()

	ctx.BindBuffer(gl.ARRAY_BUFFER, gl.Buffer(b))
	p.logError()
	ctx.DeleteBuffer(gl.Buffer(b))
	p.logError()
}

func (p *glPainter) glDrawTexture(texture Texture, alpha float32) {
	ctx := p.glctx()

	p.glctx().UseProgram(gl.Program(p.program))

	// here we have to choose between blending the image alpha or fading it...
	// TODO find a way to support both
	if alpha != 1.0 {
		ctx.BlendColor(0, 0, 0, alpha)
		ctx.BlendFunc(gl.CONSTANT_ALPHA, gl.ONE_MINUS_CONSTANT_ALPHA)
	} else {
		ctx.BlendFunc(gl.ONE, gl.ONE_MINUS_SRC_ALPHA)
	}
	p.logError()

	ctx.ActiveTexture(gl.TEXTURE0)
	ctx.BindTexture(gl.TEXTURE_2D, gl.Texture(texture))
	p.logError()

	ctx.DrawArrays(gl.TRIANGLE_STRIP, 0, 4)
	p.logError()
}

func (p *glPainter) glDrawLine(width float32, col color.Color, feather float32) {
	ctx := p.glctx()

	p.glctx().UseProgram(gl.Program(p.lineProgram))

	ctx.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	p.logError()

	colorUniform := ctx.GetUniformLocation(gl.Program(p.lineProgram), "color")
	r, g, b, a := col.RGBA()
	if a == 0 {
		ctx.Uniform4f(colorUniform, 0, 0, 0, 0)
	} else {
		alpha := float32(a)
		col := []float32{float32(r) / alpha, float32(g) / alpha, float32(b) / alpha, alpha / 0xffff}
		ctx.Uniform4fv(colorUniform, col)
	}
	lineWidthUniform := ctx.GetUniformLocation(gl.Program(p.lineProgram), "lineWidth")
	ctx.Uniform1f(lineWidthUniform, width)

	featherUniform := ctx.GetUniformLocation(gl.Program(p.lineProgram), "feather")
	ctx.Uniform1f(featherUniform, feather)

	ctx.DrawArrays(gl.TRIANGLES, 0, 6)
	p.logError()
}

func (p *glPainter) glCapture(width, height int32, pixels *[]uint8) {
	p.glctx().ReadPixels(*pixels, 0, 0, int(width), int(height), gl.RGBA, gl.UNSIGNED_BYTE)
	p.logError()
}

func glInit() {
}
