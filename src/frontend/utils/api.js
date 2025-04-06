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
        const response = await fetch(`${API_BASE_URL}/products`);
        if (!response.ok) {
            throw new Error('Erro ao listar produtos');
        }
        return response.json();
    },

    async get(id) {
        const response = await fetch(`${API_BASE_URL}/products/${id}`);
        if (!response.ok) {
            throw new Error('Erro ao buscar produto');
        }
        return response.json();
    },

    async create(formData) {
        const response = await fetch(`${API_BASE_URL}/products`, {
            method: 'POST',
            body: formData,
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('token')}`
            }
        });

        if (!response.ok) {
            throw new Error('Erro ao criar produto');
        }
        return response.json();
    },

    async update(id, formData) {
        const response = await fetch(`${API_BASE_URL}/products/${id}`, {
            method: 'PUT',
            body: formData,
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('token')}`
            }
        });

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.message || 'Erro ao atualizar produto');
        }
        return response.json();
    },

    async delete(id) {
        const response = await fetch(`${API_BASE_URL}/products/${id}`, {
            method: 'DELETE',
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('token')}`
            }
        });

        if (!response.ok) {
            throw new Error('Erro ao excluir produto');
        }
    },

    getImage(imageName) {
        return `${API_BASE_URL}/images/${imageName}`;
    }
};

// User API functions
export const userAPI = {
    async list() {
        return fetchAPI('/users', {
            method: 'GET',
        });
    },

    async get(id) {
        return fetchAPI(`/users/${id}`, {
            method: 'GET',
        });
    },

    async update(id, data) {
        return fetchAPI(`/users/${id}`, {
            method: 'PUT',
            body: JSON.stringify(data),
        });
    },

    async delete(id) {
        return fetchAPI(`/users/${id}`, {
            method: 'DELETE',
        });
    },
};

// Navigation utility
export function updateNavigation() {
    const nav = document.querySelector('nav');
    if (!nav) return;

    const token = localStorage.getItem('token');
    let navHTML = '';

    if (token) {
        navHTML = `
            <a href="index.html">Home</a>
            <a href="products.html">Gerenciar Produtos</a>
            <a href="users.html">Gerenciar Usuários</a>
            <button id="logoutButton">Sair</button>
        `;
    } else {
        navHTML = `
            <a href="index.html">Home</a>
            <a href="login.html">Entrar</a>
            <a href="register.html">Cadastrar</a>
        `;
    }

    nav.innerHTML = navHTML;

    // Adiciona o event listener para o botão de logout
    const logoutButton = document.getElementById('logoutButton');
    if (logoutButton) {
        logoutButton.addEventListener('click', () => {
            localStorage.removeItem('token');
            window.location.href = 'index.html';
        });
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