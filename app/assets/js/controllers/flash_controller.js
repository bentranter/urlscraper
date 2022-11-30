import { Controller } from "@hotwired/stimulus"

/**
 * FIVE_SECONDS is 5000ms.
 */
const FIVE_SECONDS = 5 * 1000

export default class extends Controller {
  static get targets() {
    return [
      // close is the "close" button.
      "close"
    ]
  }

  connect = () => {
    this.element.classList.remove("translate-y-2", "opacity-0", "sm:translate-y-0", "sm:translate-x-2")
    this.element.classList.add("translate-y-0", "opacity-100", "sm:translate-x-0")

    // After the entrance animation plays, update the animation transition and
    // timing for the exit animation.
    window.setTimeout(() => {
      this.element.classList.remove("duration-300", "ease-out", )
      this.element.classList.add("duration-100", "ease-in")
    }, 325)

    // Clear the notification automatically after five seconds.
    window.setTimeout(() => {
      this.close()
    }, FIVE_SECONDS)
  }

  close = () => {
    this.element.classList.remove("opacity-100")
    this.element.classList.add("opacity-0")

    window.setTimeout(() => {
      this.element.classList.add("hidden")
    }, 125)
  }

  disconnect = () => {
    this.close()
  }
}
