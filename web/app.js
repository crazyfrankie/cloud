// 全局变量
const API_BASE = 'http://localhost:8089';
let currentUserId = null;
let currentFolderId = 0; // 0 表示根目录
let currentPath = '/';
let folderStack = []; // 导航历史栈，存储 {id, name, path} 对象

// DOM 元素
const loginPage = document.getElementById('loginPage');
const registerPage = document.getElementById('registerPage');
const mainPage = document.getElementById('mainPage');
const loginForm = document.getElementById('loginForm');
const registerForm = document.getElementById('registerForm');
const showRegisterLink = document.getElementById('showRegister');
const showLoginLink = document.getElementById('showLogin');
const logoutBtn = document.getElementById('logoutBtn');
const userAvatar = document.getElementById('userAvatar');
const userDropdownMenu = document.getElementById('userDropdownMenu');
const userProfileBtn = document.getElementById('userProfileBtn');
const uploadBtn = document.getElementById('uploadBtn');
const createFolderBtn = document.getElementById('createFolderBtn');
const uploadModal = document.getElementById('uploadModal');
const folderModal = document.getElementById('folderModal');
const profileModal = document.getElementById('profileModal');
const uploadArea = document.getElementById('uploadArea');
const fileInput = document.getElementById('fileInput');
const folderForm = document.getElementById('folderForm');
const profileForm = document.getElementById('profileForm');
const changeAvatarBtn = document.getElementById('changeAvatarBtn');
const avatarInput = document.getElementById('avatarInput');
const cancelProfileBtn = document.getElementById('cancelProfileBtn');
const fileListContent = document.getElementById('fileListContent');
const currentPathSpan = document.getElementById('currentPath');

// 初始化
document.addEventListener('DOMContentLoaded', function() {
    checkAuth();
    bindEvents();
});

// 检查认证状态
function checkAuth() {
    if (hasAuthCookie()) {
        showMainPage();
        loadFolderContents(currentFolderId);
        // 加载用户信息和头像
        loadUserInfo();
    } else {
        showLoginPage();
    }
}

// 检查是否有认证cookie
function hasAuthCookie() {
    const cookies = document.cookie.split(';');
    for (let cookie of cookies) {
        const [name] = cookie.trim().split('=');
        if (name === 'cloud_access' || name === 'cloud_refresh') {
            return true;
        }
    }
    return false;
}

// 绑定事件
function bindEvents() {
    // 登录注册切换
    showRegisterLink.addEventListener('click', (e) => {
        e.preventDefault();
        showRegisterPage();
    });

    showLoginLink.addEventListener('click', (e) => {
        e.preventDefault();
        showLoginPage();
    });

    // 表单提交
    loginForm.addEventListener('submit', handleLogin);
    registerForm.addEventListener('submit', handleRegister);
    folderForm.addEventListener('submit', handleCreateFolder);
    profileForm.addEventListener('submit', handleUpdateProfile);

    // 按钮点击
    logoutBtn.addEventListener('click', handleLogout);
    userProfileBtn.addEventListener('click', showUserProfile);
    changeAvatarBtn.addEventListener('click', () => avatarInput.click());
    cancelProfileBtn.addEventListener('click', closeProfileModal);
    uploadBtn.addEventListener('click', () => uploadModal.classList.remove('hidden'));
    createFolderBtn.addEventListener('click', () => folderModal.classList.remove('hidden'));

    // 头像上传
    avatarInput.addEventListener('change', handleAvatarUpload);

    // 用户头像下拉菜单
    userAvatar.addEventListener('click', toggleUserDropdown);
    
    // 点击页面其他地方关闭下拉菜单
    document.addEventListener('click', (e) => {
        if (!e.target.closest('.user-dropdown')) {
            userDropdownMenu.classList.remove('show');
            userDropdownMenu.classList.add('hidden');
        }
    });

    // 模态框关闭
    document.querySelectorAll('.close').forEach(closeBtn => {
        closeBtn.addEventListener('click', (e) => {
            e.target.closest('.modal').classList.add('hidden');
        });
    });

    // 点击模态框外部关闭
    uploadModal.addEventListener('click', (e) => {
        if (e.target === uploadModal) {
            uploadModal.classList.add('hidden');
        }
    });

    folderModal.addEventListener('click', (e) => {
        if (e.target === folderModal) {
            folderModal.classList.add('hidden');
        }
    });

    profileModal.addEventListener('click', (e) => {
        if (e.target === profileModal) {
            profileModal.classList.add('hidden');
        }
    });

    // 文件上传
    uploadArea.addEventListener('click', () => fileInput.click());
    fileInput.addEventListener('change', handleFileSelect);

    // 拖拽上传
    uploadArea.addEventListener('dragover', (e) => {
        e.preventDefault();
        uploadArea.classList.add('dragover');
    });

    uploadArea.addEventListener('dragleave', () => {
        uploadArea.classList.remove('dragover');
    });

    uploadArea.addEventListener('drop', (e) => {
        e.preventDefault();
        uploadArea.classList.remove('dragover');
        const files = e.dataTransfer.files;
        handleFiles(files);
    });
}

// 页面切换
function showLoginPage() {
    loginPage.classList.remove('hidden');
    registerPage.classList.add('hidden');
    mainPage.classList.add('hidden');
}

function showRegisterPage() {
    loginPage.classList.add('hidden');
    registerPage.classList.remove('hidden');
    mainPage.classList.add('hidden');
}

function showMainPage() {
    loginPage.classList.add('hidden');
    registerPage.classList.add('hidden');
    mainPage.classList.remove('hidden');
}

// 处理登录
async function handleLogin(e) {
    e.preventDefault();
    const formData = new FormData(e.target);
    
    try {
        const response = await fetch(`${API_BASE}/auth/login`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            credentials: 'include', // 包含cookie
            body: JSON.stringify({
                nickname: formData.get('nickname'),
                password: formData.get('password')
            })
        });

        const result = await response.json();
        
        if (result.code === 20000) {
            // 登录成功后获取真实的用户信息
            const userInfo = await getUserInfo();
            if (userInfo) {
                // 保存并显示用户信息
                localStorage.setItem('userNickname', userInfo.nickname);
                updateUserAvatar(userInfo.avatar);
                showMainPage();
                loadFolderContents(currentFolderId);
            } else {
                // 获取用户信息失败，使用登录表单中的昵称作为备选
                localStorage.setItem('userNickname', formData.get('nickname'));
                showMainPage();
                loadFolderContents(currentFolderId);
            }
        } else {
            alert(result.msg || '登录失败');
        }
    } catch (error) {
        console.error('Login error:', error);
        alert('登录失败，请检查网络连接');
    }
}

async function getUserInfo() {
    try {
        const response = await fetch(`${API_BASE}/user`, {
            method: 'GET',
            credentials: 'include'
        });

        const result = await response.json();

        if (result.code === 20000) {
            return result.data; // 返回用户信息，包含nickname等
        } else {
            return null;
        }
    } catch (error) {
        console.error('Get user info error:', error);
        return null;
    }
}

// 加载用户信息和头像
async function loadUserInfo() {
    const userInfo = await getUserInfo();
    if (userInfo) {
        localStorage.setItem('userNickname', userInfo.nickname);
        updateUserAvatar(userInfo.avatar);
    }
}

// 更新用户头像
function updateUserAvatar(avatarUrl) {
    if (userAvatar && avatarUrl) {
        userAvatar.src = avatarUrl;
    }
}

// 切换用户下拉菜单显示状态
function toggleUserDropdown(e) {
    e.stopPropagation();
    const isHidden = userDropdownMenu.classList.contains('hidden');
    
    if (isHidden) {
        userDropdownMenu.classList.remove('hidden');
        userDropdownMenu.classList.add('show');
    } else {
        userDropdownMenu.classList.add('hidden');
        userDropdownMenu.classList.remove('show');
    }
}

// 显示用户个人信息
async function showUserProfile() {
    // 关闭下拉菜单
    userDropdownMenu.classList.add('hidden');
    userDropdownMenu.classList.remove('show');
    
    // 获取最新的用户信息
    const userInfo = await getUserInfo();
    if (userInfo) {
        // 填充个人信息表单
        document.getElementById('profileAvatar').src = userInfo.avatar || 'http://localhost:9000/cloud-user/default.jpg';
        document.getElementById('profileNickname').value = userInfo.nickname || '';
        document.getElementById('profileBirthday').value = userInfo.birthday ? userInfo.birthday.split(' ')[0] : '';
        document.getElementById('registerTime').textContent = formatDate(userInfo.utime);
        
        // 显示个人信息模态框
        profileModal.classList.remove('hidden');
    } else {
        alert('获取用户信息失败');
    }
}

// 关闭个人信息模态框
function closeProfileModal() {
    profileModal.classList.add('hidden');
}

// 处理个人信息更新
async function handleUpdateProfile(e) {
    e.preventDefault();
    const formData = new FormData(e.target);
    
    try {
        const response = await fetch(`${API_BASE}/user/update/info`, {
            method: 'PATCH',
            headers: {
                'Content-Type': 'application/json',
            },
            credentials: 'include',
            body: JSON.stringify({
                nickname: formData.get('nickname'),
                birthday: formData.get('birthday')
            })
        });

        const result = await response.json();
        
        if (result.code === 20000) {
            alert('个人信息更新成功！');
            // 更新本地存储的昵称
            localStorage.setItem('userNickname', formData.get('nickname'));
            // 关闭模态框
            closeProfileModal();
            // 重新加载用户信息更新头像显示
            loadUserInfo();
        } else {
            alert(result.msg || '更新失败');
            if (result.code === 40001) {
                clearAuthInfo();
                showLoginPage();
            }
        }
    } catch (error) {
        console.error('Update profile error:', error);
        alert('更新失败，请检查网络连接');
    }
}

// 处理头像上传
async function handleAvatarUpload(e) {
    const file = e.target.files[0];
    if (!file) return;
    
    // 检查文件类型
    if (!file.type.startsWith('image/')) {
        alert('请选择图片文件！');
        return;
    }
    
    // 检查文件大小（限制5MB）
    if (file.size > 5 * 1024 * 1024) {
        alert('图片文件不能超过5MB！');
        return;
    }
    
    try {
        // 1. 获取头像上传的预签名URL
        const presignResponse = await getAvatarPresignedUrl(file.name);
        if (!presignResponse.success) {
            throw new Error(presignResponse.error);
        }

        // 2. 上传头像到MinIO
        const uploadResponse = await uploadToMinio(presignResponse.data.presignedUrl, file);
        if (!uploadResponse) {
            throw new Error('上传头像失败');
        }

        // 3. 更新用户头像
        const updateResponse = await updateUserAvatarInDB(presignResponse.data.objectKey);
        if (!updateResponse.success) {
            throw new Error(updateResponse.error);
        }

        // 4. 更新界面显示
        const newAvatarUrl = presignResponse.data.presignedUrl.split('?')[0]; // 去掉查询参数
        document.getElementById('profileAvatar').src = newAvatarUrl;
        updateUserAvatarDisplay(newAvatarUrl);
        
        alert('头像更新成功！');
    } catch (error) {
        console.error('Avatar upload error:', error);
        alert(`头像上传失败: ${error.message}`);
    }
}

// 获取头像上传的预签名URL
async function getAvatarPresignedUrl(filename) {
    try {
        // 生成唯一的文件名
        const timestamp = Date.now();
        const extension = filename.split('.').pop();
        const uniqueFilename = `avatar_${timestamp}.${extension}`;
        
        const formData = new FormData();
        formData.append('filename', uniqueFilename);

        const response = await fetch(`${API_BASE}/storage/presign/avatar`, {
            method: 'POST',
            credentials: 'include',
            body: formData
        });

        const result = await response.json();
        
        if (result.code === 20000) {
            return { 
                success: true, 
                data: {
                    presignedUrl: result.data.presignedUrl,
                    objectKey: result.data.objectKey || uniqueFilename
                }
            };
        } else {
            if (result.code === 40001) {
                clearAuthInfo();
                showLoginPage();
            }
            return { success: false, error: result.msg || '获取上传链接失败' };
        }
    } catch (error) {
        return { success: false, error: '网络错误' };
    }
}

// 更新用户头像到数据库
async function updateUserAvatarInDB(objectKey) {
    try {
        const formData = new FormData();
        formData.append('object', objectKey);

        const response = await fetch(`${API_BASE}/user/update/avatar`, {
            method: 'PATCH',
            credentials: 'include',
            body: formData
        });

        const result = await response.json();
        
        if (result.code === 20000) {
            return { success: true };
        } else {
            if (result.code === 40001) {
                clearAuthInfo();
                showLoginPage();
            }
            return { success: false, error: result.msg || '更新头像失败' };
        }
    } catch (error) {
        return { success: false, error: '网络错误' };
    }
}

// 更新用户头像显示
function updateUserAvatarDisplay(avatarUrl) {
    if (userAvatar && avatarUrl) {
        userAvatar.src = avatarUrl;
    }
}

// 处理注册
async function handleRegister(e) {
    e.preventDefault();
    const formData = new FormData(e.target);
    
    try {
        const response = await fetch(`${API_BASE}/user/register`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                password: formData.get('password'),
                nickname: formData.get('nickname')
            })
        });

        const result = await response.json();
        
        if (result.code === 20000) {
            alert('注册成功，请登录');
            showLoginPage();
        } else {
            alert(result.msg || '注册失败');
        }
    } catch (error) {
        console.error('Register error:', error);
        alert('注册失败，请检查网络连接');
    }
}

// 处理登出
async function handleLogout() {
    try {
        // 调用后端登出API以清除cookie
        await fetch(`${API_BASE}/auth/logout`, {
            method: 'GET', // 后端使用GET方法
            credentials: 'include'
        });
    } catch (error) {
        console.error('Logout error:', error);
    }
    
    // 清除本地存储的用户信息
    clearAuthInfo();
    showLoginPage();
}

// 加载文件夹内容
async function loadFolderContents(folderId) {
    try {
        const response = await fetch(`${API_BASE}/folder/${folderId}`, {
            credentials: 'include' // 使用cookie认证
        });

        const result = await response.json();
        
        if (result.code === 20000) {
            renderFileList(result.data);
        } else {
            console.error('Failed to load folder contents:', result.msg);
            // 如果是认证失败，清理认证信息并跳转到登录页
            if (result.code === 40001) {
                clearAuthInfo();
                showLoginPage();
            } else {
                alert(`加载文件夹内容失败: ${result.msg}`);
            }
        }
    } catch (error) {
        console.error('Error loading folder contents:', error);
        alert('加载文件夹内容失败，请检查网络连接');
    }
}

// 渲染文件列表
function renderFileList(data) {
    const { files = [], folders = [] } = data;
    fileListContent.innerHTML = '';

    // 如果不在根目录，显示返回上级目录
    if (currentFolderId !== 0) {
        const backItem = createBackItem();
        fileListContent.appendChild(backItem);
    }

    // 渲染文件夹
    folders.forEach(folder => {
        const folderItem = createFolderItem(folder);
        fileListContent.appendChild(folderItem);
    });

    // 渲染文件
    files.forEach(file => {
        const fileItem = createFileItem(file);
        fileListContent.appendChild(fileItem);
    });

    if (files.length === 0 && folders.length === 0 && currentFolderId === 0) {
        fileListContent.innerHTML = '<div style="padding: 40px; text-align: center; color: #666;">暂无文件，开始上传您的第一个文件吧！</div>';
    }
}

// 创建返回上级目录项
function createBackItem() {
    const div = document.createElement('div');
    div.className = 'file-item';
    div.innerHTML = `
        <div class="file-name">
            <span class="file-icon">📁</span>
            <span>.. 返回上级目录</span>
        </div>
        <div class="file-size">-</div>
        <div class="file-date">-</div>
        <div class="file-actions"></div>
    `;
    
    div.addEventListener('click', () => {
        goBackToParent();
    });
    
    return div;
}

// 创建文件夹项
function createFolderItem(folder) {
    const div = document.createElement('div');
    div.className = 'file-item';
    div.innerHTML = `
        <div class="file-name">
            <span class="file-icon folder-icon">📁</span>
            <span>${folder.name}</span>
        </div>
        <div class="file-size">-</div>
        <div class="file-date">${formatDate(folder.utime)}</div>
        <div class="file-actions">
            <button class="action-btn" onclick="deleteFolder(${folder.id})">删除</button>
        </div>
    `;
    
    div.addEventListener('dblclick', () => {
        enterFolder(folder);
    });
    
    return div;
}

// 创建文件项
function createFileItem(file) {
    const div = document.createElement('div');
    div.className = 'file-item';
    div.innerHTML = `
        <div class="file-name">
            <span class="file-icon file">📄</span>
            <span>${file.name}</span>
        </div>
        <div class="file-size">${formatFileSize(file.size)}</div>
        <div class="file-date">${formatDate(file.utime)}</div>
        <div class="file-actions">
            <button class="action-btn" onclick="preWatchFile('${file.url}', '${file.name}')">预览</button>
            <button class="action-btn" onclick="deleteFile(${file.id})">删除</button>
        </div>
    `;
    
    return div;
}

// 格式化文件大小
function formatFileSize(bytes) {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

// 格式化日期
function formatDate(timestamp) {
    const date = new Date(timestamp * 1000);
    return date.toLocaleDateString('zh-CN') + ' ' + date.toLocaleTimeString('zh-CN');
}

// 处理创建文件夹
async function handleCreateFolder(e) {
    e.preventDefault();
    const formData = new FormData(e.target);
    const folderName = formData.get('folderName');
    
    try {
        const response = await fetch(`${API_BASE}/folder`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            credentials: 'include', // 使用cookie认证
            body: JSON.stringify({
                name: folderName,
                parentId: currentFolderId
            })
        });

        const result = await response.json();
        
        if (result.code === 20000) {
            folderModal.classList.add('hidden');
            document.getElementById('folderName').value = '';
            loadFolderContents(currentFolderId);
        } else {
            alert(result.msg || '创建文件夹失败');
            // 如果是认证失败，清理认证信息并跳转到登录页
            if (result.code === 40001) {
                clearAuthInfo();
                showLoginPage();
            }
        }
    } catch (error) {
        console.error('Create folder error:', error);
        alert('创建文件夹失败，请检查网络连接');
    }
}

// 处理文件选择
function handleFileSelect(e) {
    const files = e.target.files;
    handleFiles(files);
}

// 处理文件上传
async function handleFiles(files) {
    for (let i = 0; i < files.length; i++) {
        await uploadFile(files[i]);
    }
    uploadModal.classList.add('hidden');
    loadFolderContents(currentFolderId);
}

// 上传单个文件
async function uploadFile(file) {
    try {
        // 1. 获取预签名URL
        const presignResponse = await getPresignedUrl(file.name);
        if (!presignResponse.success) {
            throw new Error(presignResponse.error);
        }

        // 2. 直接上传到MinIO
        const uploadResponse = await uploadToMinio(presignResponse.data.presignedUrl, file);
        if (!uploadResponse) {
            throw new Error('上传文件到存储失败');
        }

        // 3. 保存文件元数据
        const metadataResponse = await saveFileMetadata(file, presignResponse.data.presignedUrl);
        if (!metadataResponse.success) {
            throw new Error(metadataResponse.error);
        }

        console.log('文件上传成功:', file.name);
    } catch (error) {
        console.error('文件上传失败:', error);
        alert(`文件 ${file.name} 上传失败: ${error.message}`);
    }
}

// 获取预签名URL
async function getPresignedUrl(filename) {
    try {
        const formData = new FormData();
        formData.append('filename', filename);

        const response = await fetch(`${API_BASE}/storage/presign/file`, {
            method: 'POST',
            credentials: 'include', // 使用cookie认证
            body: formData
        });

        const result = await response.json();
        
        if (result.code === 20000) {
            return { success: true, data: result.data };
        } else {
            // 如果是认证失败，清理认证信息并跳转到登录页
            if (result.code === 40001) {
                clearAuthInfo();
                showLoginPage();
            }
            return { success: false, error: result.msg || '获取预签名URL失败' };
        }
    } catch (error) {
        return { success: false, error: '网络错误' };
    }
}

// 上传到MinIO
async function uploadToMinio(presignedUrl, file) {
    try {
        const response = await fetch(presignedUrl, {
            method: 'PUT',
            body: file,
            headers: {
                'Content-Type': file.type
            }
        });

        return response.ok;
    } catch (error) {
        console.error('Upload to MinIO error:', error);
        return false;
    }
}

// 保存文件元数据
async function saveFileMetadata(file, presignedUrl) {
    try {
        // 从预签名URL中提取实际的对象URL（去掉查询参数）
        const url = new URL(presignedUrl);
        const objectUrl = `${url.protocol}//${url.host}${url.pathname}`;
        
        const response = await fetch(`${API_BASE}/file/upload`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            credentials: 'include', // 使用cookie认证
            body: JSON.stringify({
                name: file.name,
                size: file.size,
                folderId: currentFolderId,
                url: objectUrl,
                version: 1,
                deviceId: 'web-client'
            })
        });

        const result = await response.json();
        
        if (result.code === 20000) {
            return { success: true, data: result.data };
        } else {
            // 如果是认证失败，清理认证信息并跳转到登录页
            if (result.code === 40001) {
                clearAuthInfo();
                showLoginPage();
            }
            return { success: false, error: result.msg || '保存文件元数据失败' };
        }
    } catch (error) {
        return { success: false, error: '网络错误' };
    }
}

// 预览文件
function preWatchFile(url, filename) {
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
}

// 删除文件
async function deleteFile(fileId) {
    if (!confirm('确定要删除这个文件吗？')) {
        return;
    }
    
    // TODO: 实现删除文件API
    alert('删除文件功能待实现');
}

// 删除文件夹
async function deleteFolder(folderId) {
    if (!confirm('确定要删除这个文件夹吗？')) {
        return;
    }
    
    // TODO: 实现删除文件夹API
    alert('删除文件夹功能待实现');
}

// 进入文件夹
function enterFolder(folder) {
    // 将当前文件夹加入导航栈
    folderStack.push({
        id: currentFolderId,
        name: currentFolderId === 0 ? '/' : getCurrentFolderName(),
        path: currentPath
    });
    
    // 更新当前文件夹信息
    currentFolderId = folder.id;
    currentPath = folder.path || (currentPath === '/' ? `/${folder.name}` : `${currentPath}/${folder.name}`);
    
    // 更新UI显示
    if (currentPathSpan) {
        currentPathSpan.textContent = currentPath;
    }
    
    // 加载新文件夹的内容
    loadFolderContents(currentFolderId);
}

// 返回上级目录
function goBackToParent() {
    if (folderStack.length > 0) {
        // 从导航栈中弹出上级目录
        const parent = folderStack.pop();
        currentFolderId = parent.id;
        currentPath = parent.path;
        
        // 更新UI显示
        if (currentPathSpan) {
            currentPathSpan.textContent = currentPath;
        }
        
        // 加载父目录内容
        loadFolderContents(currentFolderId);
    }
}

// 获取当前文件夹名称
function getCurrentFolderName() {
    const pathParts = currentPath.split('/').filter(part => part);
    return pathParts.length > 0 ? pathParts[pathParts.length - 1] : '/';
}

// 清理认证信息
function clearAuthInfo() {
    currentUserId = null;
    localStorage.removeItem('userNickname');
    // 清理所有相关cookie
    document.cookie = 'cloud_access=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;';
    document.cookie = 'cloud_refresh=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;';
}
