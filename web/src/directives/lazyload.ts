import type { Directive, DirectiveBinding } from 'vue'

interface LazyloadElement extends HTMLElement {
  _observer?: IntersectionObserver
  _src?: string
}

/**
 * 图片懒加载指令
 */
export const lazyload: Directive<LazyloadElement, string> = {
  mounted(el, binding: DirectiveBinding<string>) {
    const observer = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            const target = entry.target as LazyloadElement
            if (binding.value) {
              // @ts-ignore - setting src on img element
              target.src = binding.value
              target._src = binding.value
            }
            observer.unobserve(target)
          }
        })
      },
      {
        rootMargin: '50px',
        threshold: 0.01
      }
    )

    observer.observe(el)
    el._observer = observer
  },

  updated(el, binding: DirectiveBinding<string>) {
    if (binding.value !== el._src) {
      el._src = binding.value
      if (el._observer) {
        el._observer.disconnect()
      }
      el._observer = new IntersectionObserver(
        (entries) => {
          entries.forEach((entry) => {
            if (entry.isIntersecting) {
              const target = entry.target as LazyloadElement
              if (binding.value) {
                // @ts-ignore - setting src on img element
                target.src = binding.value
              }
              target._observer?.unobserve(target)
            }
          })
        },
        {
          rootMargin: '50px',
          threshold: 0.01
        }
      )
      el._observer.observe(el)
    }
  },

  unmounted(el) {
    if (el._observer) {
      el._observer.disconnect()
    }
  }
}

/**
 * 注册懒加载指令
 */
export function setupLazyload(app: any) {
  app.directive('lazyload', lazyload)
}
