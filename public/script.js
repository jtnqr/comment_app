const { createApp } = Vue

createApp({
    data() {
        return {
            isDarkMode: localStorage.getItem("theme") === "dark",
            newComment: '',
            comments: [],
            isSubmitting: false,
            ws: null,
            connectionAttempts: 0,
            MAX_CONNECTION_ATTEMPTS: 5,
            RECONNECT_DELAY: 5000
        }
    },
    methods: {
        formatDate(dateString) {
            return new Date(dateString).toLocaleString('id-ID', {
                day: 'numeric',
                month: 'long',
                year: 'numeric',
                hour: '2-digit',
                minute: '2-digit',
                second: '2-digit'
            })
        },
        connectWebSocket() {
            if (this.ws && this.ws.readyState === WebSocket.OPEN) return

            if (this.connectionAttempts >= this.MAX_CONNECTION_ATTEMPTS) {
                console.error("Max connection attempts reached")
                return
            }

            this.ws = new WebSocket(`${window.location.origin}/ws`)

            this.ws.onopen = () => {
                this.connectionAttempts = 0
                console.log("WebSocket connected")
            }

            this.ws.onclose = () => {
                this.connectionAttempts++
                console.warn(`WebSocket disconnected. Attempt ${this.connectionAttempts}`)

                if (this.connectionAttempts < this.MAX_CONNECTION_ATTEMPTS) {
                    setTimeout(() => this.connectWebSocket(), this.RECONNECT_DELAY)
                }
            }

            this.ws.onerror = (error) => {
                console.error("WebSocket error:", error)
            }

            this.ws.onmessage = (event) => {
                try {
                    const comment = JSON.parse(event.data)
                    this.comments.unshift({
                        id: comment.id,
                        content: comment.content,
                        created_on: this.formatDate(comment.created_on)
                    })
                } catch (error) {
                    console.error("Message parsing error:", error)
                }
            }
        },
        submitComment() {
            const trimmedComment = this.newComment.trim()

            if (!trimmedComment || this.isSubmitting) return

            if (trimmedComment.length > 500) {
                alert("Comment too long. Maximum 500 characters.")
                return
            }

            this.isSubmitting = true

            try {
                if (this.ws?.readyState === WebSocket.OPEN) {
                    this.ws.send(JSON.stringify({ content: trimmedComment }))
                    this.newComment = ''

                    this.$nextTick(() => {
                        const commentsContainer = document.querySelector('.overflow-y-auto')
                        if (commentsContainer) {
                            commentsContainer.scrollTop = 0
                        }
                    })
                } else {
                    console.warn("WebSocket not open. Cannot send message.")
                }
            } catch (error) {
                console.error("Comment submission error:", error)
            } finally {
                this.isSubmitting = false
            }
        },
        toggleDarkMode() {
            this.isDarkMode = !this.isDarkMode
            document.documentElement.classList.toggle("dark", this.isDarkMode)
            localStorage.setItem("theme", this.isDarkMode ? "dark" : "light")
        },
        handleSystemThemeChange(e) {
            if (!localStorage.getItem("theme")) {
                this.isDarkMode = e.matches
                document.documentElement.classList.toggle("dark", this.isDarkMode)
            }
        }
    },
    mounted() {
        const savedTheme = localStorage.getItem("theme")
        const systemPrefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches

        if (savedTheme === "dark") {
            this.isDarkMode = true
            document.documentElement.classList.add("dark")
        } else if (savedTheme === "light") {
            this.isDarkMode = false
            document.documentElement.classList.remove("dark")
        } else if (systemPrefersDark) {
            this.isDarkMode = true
            document.documentElement.classList.add("dark")
        }

        const darkModeMediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
        darkModeMediaQuery.addEventListener('change', this.handleSystemThemeChange)

        this.connectWebSocket()
    },
    beforeDestroy() {
        const darkModeMediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
        darkModeMediaQuery.removeEventListener('change', this.handleSystemThemeChange)
        this.ws?.close()
    }
}).mount("#app")
