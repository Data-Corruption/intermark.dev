package captcha

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/altcha-org/altcha-lib-go"
	"github.com/go-chi/chi/v5"
)

const ENV_ALTCHA_HMAC_KEY = "ALTCHA_HMAC_KEY"

var hmacKey string

func Setup(r *chi.Mux) {
	hmacKey = os.Getenv(ENV_ALTCHA_HMAC_KEY)
	if hmacKey == "" {
		return
	}
	r.Get("/captcha-challenge", func(w http.ResponseWriter, r *http.Request) {
		challenge, err := altcha.CreateChallenge(altcha.ChallengeOptions{
			HMACKey:   hmacKey,
			MaxNumber: 100000,
		})
		if err != nil {
			http.Error(w, "failed to create challenge", http.StatusInternalServerError)
			return
		}
		// return JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(challenge)
	})
}

// CheckCaptcha verifies the provided payload using the HMAC key.
//
// Example Usage - Protecting contact info from spam bots:
//
// Frontend:
//
//	<!-- Form to reveal protected content -->
//	<form id="contact-form">
//	  <altcha-widget id="altcha" class="hidden" challengeurl="/captcha-challenge" floating floatingpersist="focus"></altcha-widget>
//	  <button type="submit">Reveal Contact Information</button>
//	</form>
//
//	<!-- Output div to hold protected content -->
//	<div id="contact-info"></div>
//
//	<script>
//	  document.querySelector('#altcha').addEventListener('statechange', async (ev) => {
//	    // after altcha calculates challenge answer, include it in request for protected content
//	    if (ev.detail.state === 'verified') {
//	      const res = await fetch("/contact-info", {
//	        method: "POST",
//	        headers: { "Content-Type": "application/json" },
//	        body: JSON.stringify({ altcha: ev.detail.payload }),
//	      });
//	      if (!res.ok) {
//	        alert("CAPTCHA failed, please refresh the page and try again.");
//	        return;
//	      }
//	      const data = await res.json();
//	      document.getElementById("contact-info").innerHTML = data.contact_info;
//	      document.getElementById("altcha").reset();
//	      document.getElementById("contact-form").classList.add("hidden");
//	    }
//	  });
//	  document.getElementById("contact-form").addEventListener("submit", (ev) => {
//	    ev.preventDefault();
//	    document.getElementById("altcha").verify();
//	  });
//	</script>
//
// Backend (in Router New()):
//
//	r.Router.Get("/contact-info", func(res http.ResponseWriter, req *http.Request) {
//		var reqJSON struct {
//			Altcha string `json:"altcha"`
//		}
//		if err := json.NewDecoder(req.Body).Decode(&reqJSON); err != nil {
//			http.Error(res, "Invalid request", http.StatusBadRequest)
//			return
//		}
//		if ok := CheckCaptcha(req.Altcha); !ok {
//			http.Error(res, "CAPTCHA verification failed", http.StatusForbidden)
//			return
//		}
//		// CAPTCHA verified successfully
//		contact_info := "person@example.com"
//		json.NewEncoder(res).Encode(map[string]string{"contact_info": contact_info})
//	})
func CheckCaptcha(payload string) bool {
	if hmacKey == "" {
		return false // change for desired behavior
	}
	valid, err := altcha.VerifySolution(payload, hmacKey, true)
	if err != nil || !valid {
		return false
	}
	return true
}
