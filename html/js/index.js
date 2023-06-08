const eventSource = new EventSource('/sse');

const messagesDiv = document.getElementById('messages');

const username = prompt('Enter your username:');

const messageInputForm = document.getElementById('message-form');

messageInputForm.addEventListener('submit', async (event) => {
  event.preventDefault();
  
  const messageInput = document.getElementById('new-message').value;

  await fetch('/message', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      username,
      message: messageInput
    })
  })

  document.getElementById('new-message').value = '';
})


eventSource.addEventListener('message', function (event) {
  const messageContainer = document.createElement("div");
  messageContainer.className = 'message-container';

  const messageWrapper = document.createElement("div");
  messageWrapper.className = 'message-and-username';

  const usernameElement = document.createElement("span");
  usernameElement.innerText = data.username;
  usernameElement.className = 'username';

  const messageElement = document.createElement("span");
  messageElement.innerText = data.text;
  messageElement.className = 'message';

  const timeElement = document.createElement("span");
  timeElement.innerText = new Date(data.createdAt).toLocaleDateString();
  timeElement.className = 'time';

  messageWrapper.appendChild(usernameElement);
  messageWrapper.appendChild(messageElement);
  messageContainer.appendChild(messageWrapper);
  messageContainer.appendChild(timeElement);


  messagesDiv.appendChild(messageContainer);
}, false);

