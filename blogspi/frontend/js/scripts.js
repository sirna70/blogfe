function showLoginPage() {
    window.location.href = 'login.html';
}

function showRegisterPage() {
    window.location.href = 'register.html';
}

document.addEventListener('DOMContentLoaded', function() {
    const loginForm = document.getElementById('loginForm');
    if (loginForm) {
        loginForm.addEventListener('submit', async function(event) {
            event.preventDefault();
            const username = loginForm.username.value;
            const password = loginForm.password.value;

            const response = await fetch('http://localhost:9090/login', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ username, password })
            });

            const result = await response.json();
            if (result.status === 'success') {
                localStorage.setItem('token', result.token);
                window.location.href = result.role === 'admin' ? 'admin_dashboard.html' : 'user_dashboard.html';
            } else {
                alert(result.message);
            }
        });
    }

    const registerForm = document.getElementById('registerForm');
    if (registerForm) {
        registerForm.addEventListener('submit', async function(event) {
            event.preventDefault();
            const username = registerForm.username.value;
            const password = registerForm.password.value;
            const role = registerForm.role.value;

            const response = await fetch('http://localhost:9090/register', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ username, password, role })
            });

            const result = await response.json();
            if (result.status === 'success') {
                alert('Registration successful. You can now log in.');
                window.location.href = 'login.html';
            } else {
                alert(result.message);
            }
        });
    }
});
