function handleRegister() {
    // Получаем значения полей
    const username = document.getElementById('username').value;
    const email = document.getElementById('email').value;
    const password = document.getElementById('password').value;
    const confirmPassword = document.getElementById('confirm-password').value;

    // Простая валидация пароля
    if (password !== confirmPassword) {
        alert('Пароли не совпадают!');
        return;
    }

    // Проверяем, что поля не пустые
    if (!username || !email || !password || !confirmPassword) {
        alert('Пожалуйста, заполните все поля!');
        return;
    }

    // Выводим данные в консоль (или отправляем их на сервер)
    console.log('Имя пользователя:', username);
    console.log('Электронная почта:', email);
    console.log('Пароль:', password);

    // Выводим сообщение об успешной регистрации
    alert('Регистрация прошла успешно!');
}


function handleLogin() {
    // Получаем значения полей
    const username = document.getElementById('username').value;
    const email = document.getElementById('email').value;
    const password = document.getElementById('password').value;
    const confirmPassword = document.getElementById('confirm-password').value;

    // Простая валидация пароля
    if (password !== confirmPassword) {
        alert('Пароли не совпадают!');
        return;
    }

    // Проверяем, что поля не пустые
    if (!username || !email || !password || !confirmPassword) {
        alert('Пожалуйста, заполните все поля!');
        return;
    }

    // Выводим данные в консоль (или отправляем их на сервер)
    console.log('Имя пользователя:', username);
    console.log('Электронная почта:', email);
    console.log('Пароль:', password);

    // Выводим сообщение об успешной регистрации
    alert('Регистрация прошла успешно!');
}