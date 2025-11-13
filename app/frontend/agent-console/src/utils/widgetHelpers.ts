export const widgetShapes = [
  { value: 'rounded', label: 'Rounded', desc: 'Friendly and approachable', preview: '' },
  { value: 'square', label: 'Square', desc: 'Clean and professional', preview: '' },
  // { value: 'minimal', label: 'Minimal', desc: 'Ultra-clean design', preview: '' },
  // { value: 'professional', label: 'Professional', desc: 'Enterprise-grade', preview: '' },
  // { value: 'modern', label: 'Modern', desc: 'Contemporary style', preview: '' },
  // { value: 'classic', label: 'Classic', desc: 'Traditional design', preview: '' }
] as const

export const bubbleStyles = [
  { value: 'modern', label: 'Modern', desc: 'Sleek' },
  { value: 'classic', label: 'Classic', desc: 'Traditional' },
  { value: 'minimal', label: 'Minimal', desc: 'Simple' },
  { value: 'bot', label: 'Bot', desc: 'Bot' }
] as const

export const getWidgetButtonSize = (size: string) => {
  const sizes = {
    small: 'h-12 w-12',
    medium: 'h-14 w-14',
    large: 'h-16 w-16'
  }
  return sizes[size as keyof typeof sizes] || sizes.medium
}

export const getWidgetWindowSize = (size: string) => {
  const sizes = {
    small: 'h-96 w-80',
    medium: 'h-[450px] w-96',
    large: 'h-[500px] w-[400px]'
  }
  return sizes[size as keyof typeof sizes] || sizes.medium
}

export const getBubbleStyleIconProps = (style: string) => {
  switch (style) {
    case 'modern':
      return {
        xmlns: "http://www.w3.org/2000/svg",
        width: "24",
        height: "24",
        viewBox: "0 0 24 24",
        fill: "none",
        stroke: "currentColor",
        strokeWidth: "2.5",
        strokeLinecap: "round" as const,
        strokeLinejoin: "round" as const,
        className: "lucide lucide-message-circle-more",
        children: [
          { tag: 'path', d: "M2.992 16.342a2 2 0 0 1 .094 1.167l-1.065 3.29a1 1 0 0 0 1.236 1.168l3.413-.998a2 2 0 0 1 1.099.092 10 10 0 1 0-4.777-4.719" },
          { tag: 'path', d: "M8 12h.01" },
          { tag: 'path', d: "M12 12h.01" },
          { tag: 'path', d: "M16 12h.01" }
        ]
      }
    case 'classic':
      return {
        width: "24",
        height: "24",
        viewBox: "0 0 24 24",
        fill: "currentColor",
        children: [
          { tag: 'path', d: "M12 2C6.48 2 2 6.48 2 12c0 1.54.36 2.98.97 4.29L1 23l6.71-1.97C9.02 21.64 10.46 22 12 22c5.52 0 10-4.48 10-10S17.52 2 12 2z" }
        ]
      }
    case 'minimal':
      return {
        width: "24",
        height: "24",
        viewBox: "0 0 24 24",
        fill: "currentColor",
        children: [
          { tag: 'path', d: "M4 4h16v12H5.17L4 17.17V4m0-2c-1.1 0-2 .9-2 2v18l4-4h14c1.1 0 2-.9 2-2V4c0-1.1-.9-2-2-2H4z" }
        ]
      }
    case 'bot':
      return {
        xmlns: "http://www.w3.org/2000/svg",
        width: "24",
        height: "24",
        viewBox: "0 0 24 24",
        fill: "none",
        stroke: "currentColor",
        strokeWidth: "2",
        strokeLinecap: "round" as const,
        strokeLinejoin: "round" as const,
        className: "lucide lucide-bot",
        children: [
          { tag: 'path', d: "M12 8V4H8" },
          { tag: 'rect', d: "", width: "16", height: "12", x: "4", y: "8", rx: "2" },
          { tag: 'path', d: "M2 14h2" },
          { tag: 'path', d: "M20 14h2" },
          { tag: 'path', d: "M15 13v2" },
          { tag: 'path', d: "M9 13v2" }
        ]
      }
    default:
      return {
        width: "24",
        height: "24",
        viewBox: "0 0 24 24",
        fill: "currentColor",
        children: [
          { tag: 'path', d: "M20 2H4c-1.1 0-2 .9-2 2v12c0 1.1.9 2 2 2h4l4 4 4-4h4c1.1 0 2-.9 2-2V4c0-1.1-.9-2-2-2H4z" }
        ]
      }
  }
}
