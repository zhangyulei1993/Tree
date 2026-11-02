const MAX_BYTES = 800 * 1024
const MAX_EDGE = 1440
const MIN_EDGE = 720

function loadImage(file: File): Promise<HTMLImageElement> {
  return new Promise((resolve, reject) => {
    const url = URL.createObjectURL(file)
    const image = new Image()
    image.onload = () => {
      URL.revokeObjectURL(url)
      resolve(image)
    }
    image.onerror = () => {
      URL.revokeObjectURL(url)
      reject(new Error('图片读取失败'))
    }
    image.src = url
  })
}

function canvasToBlob(canvas: HTMLCanvasElement, type: string, quality: number): Promise<Blob> {
  return new Promise((resolve, reject) => {
    canvas.toBlob((blob) => {
      if (blob) resolve(blob)
      else reject(new Error('图片压缩失败'))
    }, type, quality)
  })
}

export async function compressUploadImage(file: File): Promise<File> {
  const image = await loadImage(file)
  if (file.size <= MAX_BYTES) return file
  let scale = Math.min(1, MAX_EDGE / image.width, MAX_EDGE / image.height)
  let blob: Blob | null = null

  while (scale >= MIN_EDGE / Math.max(image.width, image.height)) {
    const canvas = document.createElement('canvas')
    canvas.width = Math.max(1, Math.round(image.width * scale))
    canvas.height = Math.max(1, Math.round(image.height * scale))
    const context = canvas.getContext('2d')
    if (!context) throw new Error('图片压缩失败')
    context.drawImage(image, 0, 0, canvas.width, canvas.height)

    for (const quality of [0.78, 0.66, 0.56, 0.48]) {
      const candidate = await canvasToBlob(canvas, 'image/webp', quality)
      blob = candidate
      if (candidate.size <= MAX_BYTES) break
    }
    if (blob && blob.size <= MAX_BYTES) break
    scale *= 0.84
  }

  // Never replace a smaller source with a larger encoded result. This also
  // covers large-dimension images whose WebP output does not reach the target.
  if (!blob || blob.size > MAX_BYTES) throw new Error('图片压缩后仍超过 800KB，请换一张更小的图片')
  if (blob.size >= file.size) return file

  const baseName = file.name.replace(/\.[^.]+$/, '') || 'launch-splash'
  return new File([blob], `${baseName}.webp`, { type: 'image/webp', lastModified: Date.now() })
}
