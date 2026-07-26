import { useEffect, useRef, useState, type ImgHTMLAttributes } from 'react'

interface LazyImageProps extends Omit<ImgHTMLAttributes<HTMLImageElement>, 'src'> {
  src?: string
  rootSelector?: string
}

export function LazyImage({ src, rootSelector, ...props }: LazyImageProps) {
  const imageRef = useRef<HTMLImageElement>(null)
  const [shouldLoad, setShouldLoad] = useState(false)
  const [isLoaded, setIsLoaded] = useState(false)

  useEffect(() => {
    setShouldLoad(false)
    setIsLoaded(false)
    if (!src) return
    const image = imageRef.current
    if (!image || !('IntersectionObserver' in window)) {
      setShouldLoad(true)
      return
    }
    const root = rootSelector ? image.closest(rootSelector) : null
    const observer = new IntersectionObserver(entries => {
      if (!entries[0]?.isIntersecting) return
      setShouldLoad(true)
      observer.disconnect()
    }, { root, rootMargin: '420px 0px' })
    observer.observe(image)
    return () => observer.disconnect()
  }, [rootSelector, src])

  const { className, onError, onLoad, ...imageProps } = props
  return <img ref={imageRef} src={shouldLoad ? src : undefined} data-src={shouldLoad ? undefined : src} loading="lazy" decoding="async" aria-busy={!isLoaded} className={`${className || ''} lazy-image${isLoaded ? ' is-loaded' : ' is-loading'}`} onLoad={event => { setIsLoaded(true); onLoad?.(event) }} onError={event => { setIsLoaded(true); onError?.(event) }} {...imageProps} />
}
