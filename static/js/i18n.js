// i18n.js - Translation Engine for Cold Storage Management
const i18n = {
    currentLang: localStorage.getItem('lang') || 'hi',
    translations: {},
    loaded: false,
    initPromise: null,

    async init() {
        // Prevent double initialization
        if (this.initPromise) {
            return this.initPromise;
        }

        this.initPromise = (async () => {
            try {
                await this.loadTranslations(this.currentLang);
                this.applyTranslations();
                this.updateLangSelector();
                this.loaded = true;
            } catch (error) {
                console.error('Failed to initialize i18n:', error);
                // Fallback - keep original text
            }
        })();

        return this.initPromise;
    },

    async loadTranslations(lang) {
        try {
            const response = await fetch(`/static/locales/${lang}.json`);
            if (!response.ok) throw new Error('Failed to load translations');
            this.translations = await response.json();
        } catch (error) {
            console.error('Error loading translations:', error);
            this.translations = {};
        }
    },

    // Get translation for a key
    t(key, fallback = null) {
        return this.translations[key] || fallback || key;
    },

    // Set language and reload
    setLanguage(lang) {
        localStorage.setItem('lang', lang);
        this.currentLang = lang;
        location.reload();
    },

    // Apply translations to all elements with data-i18n attribute
    applyTranslations() {
        // Text content
        document.querySelectorAll('[data-i18n]').forEach(el => {
            const key = el.getAttribute('data-i18n');
            const translation = this.t(key);
            if (translation !== key) {
                el.textContent = translation;
            }
        });

        // Placeholders
        document.querySelectorAll('[data-i18n-placeholder]').forEach(el => {
            const key = el.getAttribute('data-i18n-placeholder');
            const translation = this.t(key);
            if (translation !== key) {
                el.placeholder = translation;
            }
        });

        // Title attributes
        document.querySelectorAll('[data-i18n-title]').forEach(el => {
            const key = el.getAttribute('data-i18n-title');
            const translation = this.t(key);
            if (translation !== key) {
                el.title = translation;
            }
        });

        // Page title
        const titleEl = document.querySelector('[data-i18n-page-title]');
        if (titleEl) {
            const key = titleEl.getAttribute('data-i18n-page-title');
            document.title = this.t(key);
        }
    },

    // Update language selector dropdown to show current language
    updateLangSelector() {
        const selector = document.getElementById('langSelector');
        if (selector) {
            selector.value = this.currentLang;
        }
    },

    // Helper to get current language
    getLang() {
        return this.currentLang;
    },

    // Check if Hindi
    isHindi() {
        return this.currentLang === 'hi';
    }
};

// Auto-init when DOM is ready
document.addEventListener('DOMContentLoaded', () => i18n.init());
