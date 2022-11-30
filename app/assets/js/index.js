import * as Turbo from "@hotwired/turbo"
import { Application } from "@hotwired/stimulus"

import FlashController from "./controllers/flash_controller"

// Setup and start Stimulus and its controllers.
const application = Application.start()
application.register("flash", FlashController)

// Start Turbolinks.
Turbo.start()
