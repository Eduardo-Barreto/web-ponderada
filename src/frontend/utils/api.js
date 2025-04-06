const API_BASE_URL = 'http://localhost:8000/api/v1';

// Utility function to show loading state
function showLoading(element) {
    element.innerHTML = '<div class="loading">Carregando...</div>';
    element.style.display = 'block';
}

// Utility function to hide loading state
function hideLoading(element) {
    element.style.display = 'none';
}

// Utility function to show messages
export function showMessage(container, message, type = 'error') {
    container.innerHTML = `
        <div class="message ${type}">
            ${message}
        </div>
    `;

    if (type === 'success') {
        setTimeout(() => {
            container.innerHTML = '';
        }, 3000);
    }
}

// Utility function to handle API requests
async function fetchAPI(endpoint, options = {}) {
    const token = localStorage.getItem('token');
    const headers = {
        'Content-Type': 'application/json',
        ...options.headers,
    };

    if (token) {
        headers['Authorization'] = `Bearer ${token}`;
    }

    const response = await fetch(`${API_BASE_URL}${endpoint}`, {
        ...options,
        headers,
    });

    if (!response.ok) {
        const error = await response.json();
        throw new Error(error.message || 'Erro ao processar requisição');
    }

    return response.json();
}

// Auth API functions
export const authAPI = {
    async login(email, password) {
        try {
            const response = await fetch(`${API_BASE_URL}/auth/login`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ email, password }),
            });

            if (!response.ok) {
                const error = await response.json();
                throw new Error(error.message || 'Erro ao fazer login');
            }

            const data = await response.json();

            if (!data.token) {
                throw new Error('Token não encontrado na resposta');
            }

            return data;
        } catch (error) {
            console.error('Login error:', error);
            throw error;
        }
    },

    async register(name, email, password) {
        return fetchAPI('/auth/register', {
            method: 'POST',
            body: JSON.stringify({ name, email, password }),
        });
    }
};

// Product API functions
export const productAPI = {
    async list() {
        const products = await fetchAPI('/products');
        return products.map(product => ({
            ...product,
            image_url: product.image ? `${API_BASE_URL}/images/${product.image}` : null
        }));
    },

    async create(description, value, quantity, image) {
        const formData = new FormData();
        formData.append('description', description);
        formData.append('value', value);
        formData.append('quantity', quantity);
        formData.append('image', image);

        const token = localStorage.getItem('token');
        const response = await fetch(`${API_BASE_URL}/products`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${token}`,
            },
            body: formData,
        });

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.message || 'Erro ao criar produto');
        }

        const data = await response.json();
        return {
            ...data,
            image_url: data.image ? `${API_BASE_URL}/images/${data.image}` : null
        };
    }
};

// Navigation utility
export function updateNavigation() {

    const nav = document.querySelector('nav');

    if (!nav) {
        console.error('Navigation element not found in the DOM');
        return;
    }

    const token = localStorage.getItem('token');

    try {
        if (token) {
            nav.innerHTML = `
                <a href="index.html">Home</a>
                <a href="create-product.html">Criar Produto</a>
                <a href="#" id="logout">Logout</a>
            `;

            const logoutBtn = document.getElementById('logout');
            if (logoutBtn) {
                logoutBtn.addEventListener('click', (e) => {
                    e.preventDefault();
                    localStorage.removeItem('token');
                    window.location.href = 'index.html';
                });
            }
        } else {
            nav.innerHTML = `
                <a href="index.html">Home</a>
                <a href="login.html">Login</a>
                <a href="register.html">Registrar</a>
            `;
        }
    } catch (error) {
        console.error('Error updating navigation:', error);
    }
}

// Form validation utility
export function validateForm(form, fields) {
    let isValid = true;
    const errors = {};

    fields.forEach(field => {
        const input = form.querySelector(`[name="${field.name}"]`);
        if (!input) return;

        if (field.required && !input.value.trim()) {
            isValid = false;
            errors[field.name] = `${field.label} é obrigatório`;
        }

        if (field.type === 'email' && input.value && !input.value.includes('@')) {
            isValid = false;
            errors[field.name] = 'Email inválido';
        }

        if (field.type === 'number' && input.value) {
            const value = parseFloat(input.value);
            if (isNaN(value)) {
                isValid = false;
                errors[field.name] = `${field.label} deve ser um número`;
            } else if (field.min !== undefined && value < field.min) {
                isValid = false;
                errors[field.name] = `${field.label} deve ser maior que ${field.min}`;
            }
        }
    });

    return { isValid, errors };
} 