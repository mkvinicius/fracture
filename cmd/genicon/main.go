package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
)

func main() {
	os.MkdirAll("assets", 0755)

	// Criar imagem 512x512
	img := image.NewRGBA(image.Rect(0, 0, 512, 512))

	// Fundo escuro #0f172a
	bg := color.RGBA{15, 23, 42, 255}
	draw.Draw(img, img.Bounds(), &image.Uniform{bg}, image.Point{}, draw.Src)

	// Salvar
	f, _ := os.Create("assets/icon.png")
	defer f.Close()
	png.Encode(f, img)
}
