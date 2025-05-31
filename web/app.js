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
            <input type="checkbox" class="file-checkbox" data-file-id="${file.id}" style="margin-right: 8px;" onchange="toggleFileSelection(${file.id}, this)">
            <span class="file-icon file">📄</span>
            <span>${file.name}</span>
        </div>
        <div class="file-size">${formatFileSize(file.size)}</div>
        <div class="file-date">${formatDate(file.utime)}</div>
        <div class="file-actions">
            <button class="action-btn" onclick="preWatchFile('${file.url}', '${file.name}')">预览</button>
            <button class="action-btn" onclick="updateFile(${file.id}, '${file.name}')">更新</button>
            <button class="action-btn" onclick="showFileVersions(${file.id}, '${file.name}')">版本</button>
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

// 文件验证函数
function validateFile(file) {
    const maxSize = 10 * 1024 * 1024 * 1024; // 10GB - 支持大文件上传
    const allowedTypes = [
        // 图片类型
        'image/jpeg', 'image/jpg', 'image/png', 'image/gif', 'image/bmp', 'image/webp',
        // 文档类型
        'application/pdf', 'application/msword', 
        'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
        'application/vnd.ms-excel',
        'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
        'application/vnd.ms-powerpoint',
        'application/vnd.openxmlformats-officedocument.presentationml.presentation',
        // 文本类型
        'text/plain', 'text/csv', 'text/html', 'text/css', 'text/javascript',
        'application/json', 'application/xml',
        // 压缩文件
        'application/zip', 'application/x-rar-compressed', 'application/x-7z-compressed',
        // 音视频
        'audio/mpeg', 'audio/wav', 'audio/ogg', 'video/mp4', 'video/avi', 'video/mov',
        // 其他常用类型
        'application/octet-stream'
    ];

    // 检查文件大小
    if (file.size > maxSize) {
        return {
            valid: false,
            error: `文件大小超过限制，最大允许 ${Math.round(maxSize / 1024 / 1024 / 1024)}GB`
        };
    }

    // 检查文件类型（如果文件有类型信息）
    if (file.type && !allowedTypes.includes(file.type)) {
        return {
            valid: false,
            error: `不支持的文件类型: ${file.type}`
        };
    }

    // 检查文件名
    if (!file.name || file.name.trim() === '') {
        return {
            valid: false,
            error: '文件名不能为空'
        };
    }

    // 检查文件名长度
    if (file.name.length > 255) {
        return {
            valid: false,
            error: '文件名过长，最多255个字符'
        };
    }

    // 检查危险文件扩展名
    const dangerousExtensions = ['.exe', '.bat', '.cmd', '.scr', '.pif', '.vbs', '.js'];
    const fileName = file.name.toLowerCase();
    for (const ext of dangerousExtensions) {
        if (fileName.endsWith(ext)) {
            return {
                valid: false,
                error: `为了安全考虑，不允许上传 ${ext} 文件`
            };
        }
    }

    return { valid: true };
}

// 网络状态检查
function checkNetworkStatus() {
    return navigator.onLine;
}

// 重试机制包装器
async function withRetry(fn, maxRetries = 3, delay = 1000) {
    let lastError;
    
    for (let i = 0; i < maxRetries; i++) {
        try {
            return await fn();
        } catch (error) {
            lastError = error;
            console.warn(`尝试 ${i + 1}/${maxRetries} 失败:`, error.message);
            
            // 如果不是最后一次重试，等待后重试
            if (i < maxRetries - 1) {
                await new Promise(resolve => setTimeout(resolve, delay * (i + 1)));
                
                // 检查网络状态
                if (!checkNetworkStatus()) {
                    throw new Error('网络连接已断开，请检查网络后重试');
                }
            }
        }
    }
    
    throw lastError;
}

// 带重试的分块上传
async function uploadChunkWithRetry(presignedUrl, chunk, partNumber, maxRetries = 3) {
    let lastError;
    
    for (let attempt = 1; attempt <= maxRetries; attempt++) {
        try {
            return await uploadChunkToStorage(presignedUrl, chunk);
        } catch (error) {
            lastError = error;
            console.warn(`分块 ${partNumber} 第 ${attempt} 次上传失败:`, error.message);
            
            if (attempt < maxRetries) {
                // 指数退避重试
                const delay = Math.min(1000 * Math.pow(2, attempt - 1), 5000);
                await new Promise(resolve => setTimeout(resolve, delay));
            }
        }
    }
    
    throw new Error(`分块 ${partNumber} 上传失败（重试 ${maxRetries} 次）: ${lastError.message}`);
}

// 处理文件上传
async function handleFiles(files) {
    const uploadModal = document.getElementById('uploadModal');
    
    // 验证文件
    const validFiles = [];
    const invalidFiles = [];
    
    for (let i = 0; i < files.length; i++) {
        const validation = validateFile(files[i]);
        if (validation.valid) {
            validFiles.push(files[i]);
        } else {
            invalidFiles.push({
                file: files[i],
                error: validation.error
            });
        }
    }
    
    // 显示验证失败的文件
    if (invalidFiles.length > 0) {
        const errorMessages = invalidFiles.map(item => 
            `${item.file.name}: ${item.error}`
        ).join('\n');
        alert(`以下文件验证失败:\n${errorMessages}`);
    }
    
    // 如果没有有效文件，直接返回
    if (validFiles.length === 0) {
        return;
    }
    
    // 检查网络状态
    if (!checkNetworkStatus()) {
        alert('网络连接已断开，请检查网络后重试');
        return;
    }
    
    // 创建进度显示容器
    const progressContainer = createProgressContainer();
    uploadModal.appendChild(progressContainer);
    
    // 并发上传文件（限制并发数）
    const concurrentLimit = 3; // 最多同时上传3个文件
    const uploadPromises = [];
    const largeFileThreshold = 50 * 1024 * 1024; // 50MB
    
    for (let i = 0; i < validFiles.length; i++) {
        const progressItem = createProgressItem(validFiles[i].name);
        progressContainer.appendChild(progressItem);
        
        // 根据文件大小选择上传方式
        let uploadPromise;
        if (validFiles[i].size > largeFileThreshold) {
            uploadPromise = uploadLargeFileWithChunks(validFiles[i], progressItem);
        } else {
            uploadPromise = uploadFileWithProgress(validFiles[i], progressItem);
        }
        
        uploadPromises.push(uploadPromise);
        
        // 控制并发数
        if (uploadPromises.length >= concurrentLimit || i === validFiles.length - 1) {
            await Promise.allSettled(uploadPromises);
            uploadPromises.length = 0; // 清空数组
        }
    }
    
    // 上传完成后清理进度显示
    setTimeout(() => {
        uploadModal.classList.add('hidden');
        progressContainer.remove();
        loadFolderContents(currentFolderId);
    }, 2000);
}

// 创建进度显示容器
function createProgressContainer() {
    const container = document.createElement('div');
    container.className = 'upload-progress-container';
    container.style.cssText = `
        margin-top: 20px;
        max-height: 300px;
        overflow-y: auto;
        border: 1px solid #ddd;
        border-radius: 4px;
        padding: 10px;
    `;
    return container;
}

// 创建单个文件的进度项
function createProgressItem(filename) {
    const item = document.createElement('div');
    item.className = 'progress-item';
    item.innerHTML = `
        <div style="margin-bottom: 10px; padding: 10px; border: 1px solid #eee; border-radius: 4px;">
            <div style="font-weight: bold; margin-bottom: 5px;">${filename}</div>
            <div class="progress-bar" style="width: 100%; height: 20px; background-color: #f0f0f0; border-radius: 10px; overflow: hidden;">
                <div class="progress-fill" style="width: 0%; height: 100%; background-color: #4CAF50; transition: width 0.3s ease;"></div>
            </div>
            <div class="progress-text" style="margin-top: 5px; font-size: 12px; color: #666;">准备上传...</div>
        </div>
    `;
    return item;
}

// 带进度的文件上传
async function uploadFileWithProgress(file, progressItem) {
    const progressFill = progressItem.querySelector('.progress-fill');
    const progressText = progressItem.querySelector('.progress-text');
    
    try {
        // 1. 计算文件哈希值（带进度）
        progressText.textContent = '计算文件哈希值...';
        progressFill.style.width = '5%';
        
        const fileHash = await calculateFileHash(file, (hashProgress) => {
            const currentProgress = 5 + (hashProgress * 0.15); // 5%-20%
            progressFill.style.width = `${currentProgress}%`;
            progressText.textContent = `计算文件哈希值... ${Math.round(hashProgress)}%`;
        });

        // 2. 预上传检查（秒传检测）
        progressText.textContent = '检查文件是否已存在...';
        progressFill.style.width = '25%';
        
        const checkResponse = await preUploadCheck(file, fileHash);
        if (!checkResponse.success) {
            throw new Error(checkResponse.error);
        }

        // 检查是否需要上传（秒传功能）
        if (checkResponse.data.fileExists) {
            // 文件已存在，秒传成功
            progressText.textContent = '文件已存在，秒传成功！';
            progressFill.style.width = '100%';
            progressFill.style.backgroundColor = '#2196F3'; // 蓝色表示秒传
            
            // 显示秒传成功的提示时间稍长
            setTimeout(() => {
                progressText.textContent = '秒传完成';
            }, 500);
            return;
        }

        // 3. 文件不存在，需要上传到MinIO
        if (!checkResponse.data.presignedUrl) {
            throw new Error('未获取到上传链接');
        }

        progressText.textContent = '开始上传文件...';
        progressFill.style.width = '30%';
        
        const uploadResponse = await uploadToMinioWithProgress(
            checkResponse.data.presignedUrl, 
            file,
            (progress) => {
                const uploadProgress = 30 + (progress * 0.60); // 30%-90%
                progressFill.style.width = `${uploadProgress}%`;
                progressText.textContent = `上传中... ${Math.round(progress)}%`;
            }
        );
        
        if (!uploadResponse) {
            throw new Error('上传文件到存储失败');
        }

        // 4. 确认上传完成
        progressText.textContent = '保存文件信息...';
        progressFill.style.width = '95%';
        
        const confirmResponse = await confirmUpload(file, fileHash, checkResponse.data.presignedUrl);
        if (!confirmResponse.success) {
            throw new Error(confirmResponse.error);
        }

        // 完成
        progressFill.style.width = '100%';
        progressText.textContent = '上传成功';
        progressFill.style.backgroundColor = '#4CAF50'; // 绿色表示成功上传
        
    } catch (error) {
        console.error('文件上传失败:', error);
        progressFill.style.backgroundColor = '#f44336';
        progressText.textContent = `上传失败: ${error.message}`;
        
        // 如果是网络错误，提供重试建议
        if (error.message.includes('网络')) {
            setTimeout(() => {
                progressText.textContent += ' (建议检查网络后重试)';
            }, 1000);
        }
    }
}

// 带进度的MinIO上传
async function uploadToMinioWithProgress(presignedUrl, file, onProgress) {
    return new Promise((resolve, reject) => {
        const xhr = new XMLHttpRequest();
        
        xhr.upload.addEventListener('progress', (e) => {
            if (e.lengthComputable) {
                const percentComplete = (e.loaded / e.total) * 100;
                onProgress(percentComplete);
            }
        });
        
        xhr.addEventListener('load', () => {
            if (xhr.status >= 200 && xhr.status < 300) {
                resolve(true);
            } else {
                reject(new Error('Upload failed'));
            }
        });
        
        xhr.addEventListener('error', () => {
            reject(new Error('Upload error'));
        });
        
        xhr.open('PUT', presignedUrl);
        xhr.setRequestHeader('Content-Type', file.type);
        xhr.send(file);
    });
}

// 计算文件哈希值（使用 SHA-256，支持进度回调）
async function calculateFileHash(file, onProgress = null) {
    try {
        const chunkSize = 2 * 1024 * 1024; // 2MB chunks for progress tracking
        const chunks = Math.ceil(file.size / chunkSize);
        const crypto = window.crypto.subtle;
        
        if (file.size <= chunkSize) {
            // 小文件直接计算
            const buffer = await file.arrayBuffer();
            const hashBuffer = await crypto.digest('SHA-256', buffer);
            const hashArray = Array.from(new Uint8Array(hashBuffer));
            const hashHex = hashArray.map(b => b.toString(16).padStart(2, '0')).join('');
            if (onProgress) onProgress(100);
            return hashHex;
        }
        
        // 大文件分块计算，支持进度显示
        const hasher = await crypto.digest('SHA-256', new ArrayBuffer(0)); // 初始化
        let processedBytes = 0;
        
        // 注意：Web Crypto API 不支持流式哈希，我们需要读取整个文件
        // 这里简化为直接读取，但显示进度
        const buffer = await file.arrayBuffer();
        
        if (onProgress) {
            // 模拟分块进度
            for (let i = 0; i <= 100; i += 10) {
                if (onProgress) onProgress(i);
                await new Promise(resolve => setTimeout(resolve, 10));
            }
        }
        
        const hashBuffer = await crypto.digest('SHA-256', buffer);
        const hashArray = Array.from(new Uint8Array(hashBuffer));
        const hashHex = hashArray.map(b => b.toString(16).padStart(2, '0')).join('');
        
        if (onProgress) onProgress(100);
        return hashHex;
    } catch (error) {
        console.error('计算文件哈希失败:', error);
        throw new Error('计算文件哈希失败: ' + error.message);
    }
}

// 预上传检查（带重试机制）
async function preUploadCheck(file, fileHash) {
    return await withRetry(async () => {
        const requestData = {
            name: file.name,
            size: file.size,
            hash: fileHash,
            folderId: currentFolderId
        };

        const response = await fetch(`${API_BASE}/file/pre-upload-check`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            credentials: 'include',
            body: JSON.stringify(requestData)
        });

        if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }

        const result = await response.json();
        
        if (result.code === 20000) {
            return { 
                success: true, 
                data: result.data
            };
        } else {
            if (result.code === 40001) {
                clearAuthInfo();
                showLoginPage();
                throw new Error('登录已过期，请重新登录');
            }
            throw new Error(result.msg || '预上传检查失败');
        }
    }, 2, 1000).catch(error => {
        console.error('预上传检查失败:', error);
        return { 
            success: false, 
            error: error.message || '预上传检查失败' 
        };
    });
}

// 确认上传完成（带重试机制）
async function confirmUpload(file, fileHash, presignedUrl) {
    return await withRetry(async () => {
        // 从预签名URL提取实际的文件URL（去掉查询参数）
        const fileUrl = presignedUrl.split('?')[0];
        
        const requestData = {
            name: file.name,
            size: file.size,
            hash: fileHash,
            url: fileUrl,
            folderId: currentFolderId,
            deviceId: 'web-client'
        };

        const response = await fetch(`${API_BASE}/file/confirm-upload`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            credentials: 'include',
            body: JSON.stringify(requestData)
        });

        if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }

        const result = await response.json();
        
        if (result.code === 20000) {
            return { 
                success: true, 
                data: result.data
            };
        } else {
            if (result.code === 40001) {
                clearAuthInfo();
                showLoginPage();
                throw new Error('登录已过期，请重新登录');
            }
            throw new Error(result.msg || '确认上传失败');
        }
    }, 2, 1000).catch(error => {
        console.error('确认上传失败:', error);
        return { 
            success: false, 
            error: error.message || '确认上传失败' 
        };
    });
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

// 关闭统计信息模态框
function closeStatsModal() {
    const modal = document.querySelector('.stats-modal');
    if (modal) {
        modal.remove();
    }
}

// 在用户下拉菜单中添加统计信息按钮
function updateUserDropdownMenu() {
    const userDropdownMenu = document.getElementById('userDropdownMenu');
    if (userDropdownMenu) {
        // 检查是否已经添加了统计按钮
        if (!userDropdownMenu.querySelector('#fileStatsBtn')) {
            const statsBtn = document.createElement('a');
            statsBtn.id = 'fileStatsBtn';
            statsBtn.href = '#';
            statsBtn.textContent = '文件统计';
            statsBtn.addEventListener('click', (e) => {
                e.preventDefault();
                showFileStatistics();
                userDropdownMenu.classList.add('hidden');
            });
            
            // 插入到用户资料按钮之后
            const profileBtn = userDropdownMenu.querySelector('#userProfileBtn');
            if (profileBtn) {
                profileBtn.parentNode.insertBefore(statsBtn, profileBtn.nextSibling);
            }
        }
    }
}

// 在页面加载时更新下拉菜单
document.addEventListener('DOMContentLoaded', function() {
    // ...existing code...
    setTimeout(updateUserDropdownMenu, 1000); // 延迟执行以确保DOM已加载
});

// 删除文件
async function deleteFile(fileId) {
    if (!confirm('确定要删除这个文件吗？')) {
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/file/${fileId}`, {
            method: 'DELETE',
            credentials: 'include'
        });

        const result = await response.json();
        
        if (result.code === 20000) {
            alert('文件删除成功');
            loadFolderContents(currentFolderId);
        } else {
            alert(result.msg || '删除文件失败');
            if (result.code === 40001) {
                clearAuthInfo();
                showLoginPage();
            }
        }
    } catch (error) {
        console.error('Delete file error:', error);
        alert('删除文件失败，请检查网络连接');
    }
}

// 删除文件夹
async function deleteFolder(folderId) {
    if (!confirm('确定要删除这个文件夹吗？删除后无法恢复！')) {
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/folder/${folderId}`, {
            method: 'DELETE',
            credentials: 'include'
        });

        const result = await response.json();
        
        if (result.code === 20000) {
            alert('文件夹删除成功');
            loadFolderContents(currentFolderId);
        } else {
            alert(result.msg || '删除文件夹失败');
            if (result.code === 40001) {
                clearAuthInfo();
                showLoginPage();
            }
        }
    } catch (error) {
        console.error('Delete folder error:', error);
        alert('删除文件夹失败，请检查网络连接');
    }
}

// 批量删除文件
async function batchDeleteFiles(fileIds) {
    if (!fileIds || fileIds.length === 0) {
        alert('请选择要删除的文件');
        return;
    }

    if (!confirm(`确定要删除这 ${fileIds.length} 个文件吗？`)) {
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/file/batch-delete`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            credentials: 'include',
            body: JSON.stringify({
                fileIds: fileIds
            })
        });

        const result = await response.json();
        
        if (result.code === 20000) {
            alert('文件批量删除成功');
            loadFolderContents(currentFolderId);
        } else {
            alert(result.msg || '批量删除文件失败');
            if (result.code === 40001) {
                clearAuthInfo();
                showLoginPage();
            }
        }
    } catch (error) {
        console.error('Batch delete files error:', error);
        alert('批量删除文件失败，请检查网络连接');
    }
}

// 显示文件统计信息
async function showFileStatistics() {
    try {
        const response = await fetch(`${API_BASE}/file/stats`, {
            method: 'GET',
            credentials: 'include'
        });

        const result = await response.json();
        
        if (result.code === 20000) {
            const stats = result.data;
            
            // 创建统计信息模态框
            const modal = document.createElement('div');
            modal.className = 'stats-modal';
            modal.style.cssText = `
                position: fixed;
                top: 0;
                left: 0;
                width: 100%;
                height: 100%;
                background-color: rgba(0, 0, 0, 0.5);
                display: flex;
                justify-content: center;
                align-items: center;
                z-index: 1000;
            `;
            
            modal.innerHTML = `
                <div style="background: white; padding: 30px; border-radius: 8px; max-width: 500px; width: 90%;">
                    <h2 style="margin-top: 0; margin-bottom: 20px; text-align: center;">文件统计信息</h2>
                    <div style="line-height: 1.8;">
                        <p><strong>总文件数：</strong>${stats.totalFiles} 个</p>
                        <p><strong>总大小：</strong>${formatFileSize(stats.totalSize)}</p>
                        <p><strong>文件类型：</strong></p>
                        <div style="margin-left: 20px;">
                            ${Object.entries(stats.fileTypes || {}).map(([type, count]) => 
                                `<p>${type}: ${count} 个</p>`
                            ).join('')}
                        </div>
                    </div>
                    <div style="text-align: center; margin-top: 20px;">
                        <button onclick="closeStatsModal()" style="padding: 8px 20px; background: #007bff; color: white; border: none; border-radius: 4px; cursor: pointer;">关闭</button>
                    </div>
                </div>
            `;
            
            // 点击模态框外部关闭
            modal.addEventListener('click', (e) => {
                if (e.target === modal) {
                    closeStatsModal();
                }
            });
            
            document.body.appendChild(modal);
        } else {
            alert(result.msg || '获取统计信息失败');
            if (result.code === 40001) {
                clearAuthInfo();
                showLoginPage();
            }
        }
    } catch (error) {
        console.error('Get file stats error:', error);
        alert('获取统计信息失败，请检查网络连接');
    }
}

// 大文件分块上传功能
async function uploadLargeFileWithChunks(file, progressItem) {
    // 根据文件大小动态调整分块大小，减少分块数量
    let chunkSize;
    if (file.size <= 100 * 1024 * 1024) { // <= 100MB
        chunkSize = 5 * 1024 * 1024; // 5MB
    } else if (file.size <= 1024 * 1024 * 1024) { // <= 1GB  
        chunkSize = 10 * 1024 * 1024; // 10MB
    } else if (file.size <= 5 * 1024 * 1024 * 1024) { // <= 5GB
        chunkSize = 20 * 1024 * 1024; // 20MB
    } else {
        chunkSize = 50 * 1024 * 1024; // 50MB for very large files
    }
    
    const totalChunks = Math.ceil(file.size / chunkSize);
    
    if (totalChunks <= 1) {
        // 文件不大，使用普通上传
        return await uploadFileWithProgress(file, progressItem);
    }

    const progressFill = progressItem.querySelector('.progress-fill');
    const progressText = progressItem.querySelector('.progress-text');
    
    try {
        // 1. 计算文件哈希值
        progressText.textContent = '计算文件哈希值...';
        progressFill.style.width = '5%';
        
        const fileHash = await calculateFileHash(file, (hashProgress) => {
            const currentProgress = 5 + (hashProgress * 0.10); // 5%-15%
            progressFill.style.width = `${currentProgress}%`;
            progressText.textContent = `计算文件哈希值... ${Math.round(hashProgress)}%`;
        });

        // 2. 初始化分块上传
        progressText.textContent = '初始化分块上传...';
        progressFill.style.width = '15%';
        
        const initResponse = await initChunkedUpload(file, fileHash, totalChunks, chunkSize);
        if (!initResponse.success) {
            throw new Error(initResponse.error);
        }

        const { uploadId, chunkUrls } = initResponse.data;

        // 3. 分块上传（动态并发控制）
        progressText.textContent = '开始分块上传...';
        const uploadedChunks = [];
        
        // 根据分块数量和文件大小动态调整并发数
        let maxConcurrentChunks;
        if (totalChunks <= 10) {
            maxConcurrentChunks = 2; // 小文件用较少并发
        } else if (totalChunks <= 50) {
            maxConcurrentChunks = 3; // 中等文件
        } else if (totalChunks <= 200) {
            maxConcurrentChunks = 4; // 大文件
        } else {
            maxConcurrentChunks = 5; // 超大文件，但不超过5个并发
        }
        
        console.log(`文件分块上传信息: 
            文件大小: ${(file.size / 1024 / 1024).toFixed(2)}MB
            分块大小: ${(chunkSize / 1024 / 1024).toFixed(2)}MB  
            分块数量: ${totalChunks}
            并发数: ${maxConcurrentChunks}`);
        
        // 创建分块上传任务
        const chunkTasks = [];
        for (let chunkIndex = 0; chunkIndex < totalChunks; chunkIndex++) {
            chunkTasks.push({
                index: chunkIndex,
                partNumber: chunkIndex + 1,
                start: chunkIndex * chunkSize,
                end: Math.min((chunkIndex + 1) * chunkSize, file.size),
                url: chunkUrls.find(url => url.partNumber === chunkIndex + 1)?.presignedUrl
            });
        }
        
        // 并发上传分块
        let completedChunks = 0;
        const semaphore = new Array(maxConcurrentChunks).fill(null);
        
        await Promise.all(semaphore.map(async () => {
            while (chunkTasks.length > 0) {
                const task = chunkTasks.shift();
                if (!task) break;
                
                try {
                    const chunk = file.slice(task.start, task.end);
                    
                    // 计算分块哈希
                    const chunkHash = await calculateFileHash(chunk);
                    
                    if (!task.url) {
                        throw new Error(`未找到分块 ${task.partNumber} 的上传URL`);
                    }
                    
                    // 上传分块（带重试）
                    const etag = await uploadChunkWithRetry(task.url, chunk, task.partNumber);
                    
                    uploadedChunks.push({
                        partNumber: task.partNumber,
                        etag: etag || chunkHash
                    });
                    
                    completedChunks++;
                    
                    // 更新进度
                    const chunkProgress = (completedChunks / totalChunks) * 70; // 15%-85%
                    progressFill.style.width = `${15 + chunkProgress}%`;
                    progressText.textContent = `分块上传中... ${completedChunks}/${totalChunks}`;
                    
                } catch (error) {
                    console.error(`分块 ${task.partNumber} 上传失败:`, error);
                    // 将失败的任务重新加入队列（最多重试3次）
                    if (!task.retries) task.retries = 0;
                    if (task.retries < 3) {
                        task.retries++;
                        chunkTasks.push(task);
                    } else {
                        throw new Error(`分块 ${task.partNumber} 上传失败: ${error.message}`);
                    }
                }
            }
        }));
        
        // 按分块编号排序
        uploadedChunks.sort((a, b) => a.partNumber - b.partNumber);

        // 4. 完成分块上传
        progressText.textContent = '合并文件分块...';
        progressFill.style.width = '90%';
        
        const completeResponse = await completeChunkedUpload(uploadId, uploadedChunks);
        if (!completeResponse.success) {
            throw new Error(completeResponse.error);
        }

        // 完成
        progressFill.style.width = '100%';
        progressText.textContent = '大文件上传成功';
        progressFill.style.backgroundColor = '#4CAF50';
        
        // 记录上传完成的文件信息
        console.log('分块上传完成:', {
            fileId: completeResponse.data.fileID,
            fileUrl: completeResponse.data.fileUrl,
            message: completeResponse.data.message
        });
        
        // 刷新文件列表
        setTimeout(() => {
            loadFolderContents(currentFolderId);
        }, 1000);
        
    } catch (error) {
        console.error('分块上传失败:', error);
        progressFill.style.backgroundColor = '#f44336';
        progressText.textContent = `分块上传失败: ${error.message}`;
        
        // 尝试中止上传
        try {
            if (initResponse && initResponse.data && initResponse.data.uploadId) {
                await abortChunkedUpload(initResponse.data.uploadId);
            }
        } catch (abortError) {
            console.error('中止分块上传失败:', abortError);
        }
    }
}

// 初始化分块上传
async function initChunkedUpload(file, fileHash, totalChunks, chunkSize) {
    return await withRetry(async () => {
        const response = await fetch(`${API_BASE}/file/chunked-upload`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            credentials: 'include',
            body: JSON.stringify({
                name: file.name,
                size: file.size,
                hash: fileHash,
                folderId: currentFolderId,
                chunkSize: chunkSize,
                totalChunks: totalChunks
            })
        });

        if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }

        const result = await response.json();
        
        if (result.code === 20000) {
            return { 
                success: true, 
                data: result.data
            };
        } else {
            if (result.code === 40001) {
                clearAuthInfo();
                showLoginPage();
                throw new Error('登录已过期，请重新登录');
            }
            throw new Error(result.msg || '初始化分块上传失败');
        }
    }, 2, 1000).catch(error => {
        console.error('初始化分块上传失败:', error);
        return { 
            success: false, 
            error: error.message || '初始化分块上传失败' 
        };
    });
}

// 上传单个分块到存储
async function uploadChunkToStorage(presignedUrl, chunk) {
    return new Promise((resolve, reject) => {
        const xhr = new XMLHttpRequest();
        
        xhr.addEventListener('load', () => {
            if (xhr.status >= 200 && xhr.status < 300) {
                resolve(xhr.getResponseHeader('ETag') || 'chunk-uploaded');
            } else {
                reject(new Error('分块上传失败'));
            }
        });
        
        xhr.addEventListener('error', () => {
            reject(new Error('分块上传网络错误'));
        });
        
        xhr.open('PUT', presignedUrl);
        xhr.setRequestHeader('Content-Type', chunk.type || 'application/octet-stream');
        xhr.send(chunk);
    });
}

// 完成分块上传
async function completeChunkedUpload(uploadId, chunkETags) {
    return await withRetry(async () => {
        const response = await fetch(`${API_BASE}/file/chunked-upload/${uploadId}/complete`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            credentials: 'include',
            body: JSON.stringify({
                chunkETags: chunkETags
            })
        });

        if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }

        const result = await response.json();
        
        if (result.code === 20000) {
            return { 
                success: true, 
                data: result.data
            };
        } else {
            if (result.code === 40001) {
                clearAuthInfo();
                showLoginPage();
                throw new Error('登录已过期，请重新登录');
            }
            throw new Error(result.msg || '完成分块上传失败');
        }
    }, 2, 1000).catch(error => {
        console.error('完成分块上传失败:', error);
        return { 
            success: false, 
            error: error.message || '完成分块上传失败' 
        };
    });
}

// 中止分块上传
async function abortChunkedUpload(uploadId) {
    try {
        const response = await fetch(`${API_BASE}/file/chunked-upload/${uploadId}`, {
            method: 'DELETE',
            credentials: 'include'
        });

        const result = await response.json();
        
        if (result.code !== 20000) {
            console.error('中止分块上传失败:', result.msg);
        }
    } catch (error) {
        console.error('中止分块上传请求失败:', error);
    }
}

// 文件更新/替换功能
async function updateFile(fileId, fileName) {
    // 创建文件选择对话框
    const input = document.createElement('input');
    input.type = 'file';
    input.accept = '*/*';
    
    input.onchange = async (e) => {
        const newFile = e.target.files[0];
        if (!newFile) return;
        
        if (!confirm(`确定要用 "${newFile.name}" 替换文件 "${fileName}" 吗？`)) {
            return;
        }
        
        try {
            // 显示更新进度
            const progressContainer = createProgressContainer();
            const progressItem = createProgressItem(`更新: ${fileName}`);
            progressContainer.appendChild(progressItem);
            document.body.appendChild(progressContainer);
            
            const progressFill = progressItem.querySelector('.progress-fill');
            const progressText = progressItem.querySelector('.progress-text');
            
            // 计算新文件哈希
            progressText.textContent = '计算文件哈希值...';
            progressFill.style.width = '10%';
            
            const fileHash = await calculateFileHash(newFile, (hashProgress) => {
                const currentProgress = 10 + (hashProgress * 0.20);
                progressFill.style.width = `${currentProgress}%`;
                progressText.textContent = `计算文件哈希值... ${Math.round(hashProgress)}%`;
            });
            
            // 预上传检查
            progressText.textContent = '检查文件...';
            progressFill.style.width = '35%';
            
            const checkResponse = await preUploadCheck(newFile, fileHash);
            if (!checkResponse.success) {
                throw new Error(checkResponse.error);
            }
            
            // 如果文件需要上传
            if (!checkResponse.data.fileExists) {
                progressText.textContent = '上传新文件...';
                progressFill.style.width = '50%';
                
                const uploadResponse = await uploadToMinioWithProgress(
                    checkResponse.data.presignedUrl,
                    newFile,
                    (progress) => {
                        const uploadProgress = 50 + (progress * 0.30);
                        progressFill.style.width = `${uploadProgress}%`;
                        progressText.textContent = `上传中... ${Math.round(progress)}%`;
                    }
                );
                
                if (!uploadResponse) {
                    throw new Error('上传文件失败');
                }
                
                // 确认上传
                progressText.textContent = '确认上传...';
                progressFill.style.width = '85%';
                
                const confirmResponse = await confirmUpload(newFile, fileHash, checkResponse.data.presignedUrl);
                if (!confirmResponse.success) {
                    throw new Error(confirmResponse.error);
                }
            }
            
            // 更新文件信息
            progressText.textContent = '更新文件信息...';
            progressFill.style.width = '90%';
            
            const response = await fetch(`${API_BASE}/file/${fileId}`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json'
                },
                credentials: 'include',
                body: JSON.stringify({
                    name: newFile.name,
                    size: newFile.size,
                    hash: fileHash
                })
            });
            
            const result = await response.json();
            
            if (result.code === 20000) {
                progressFill.style.width = '100%';
                progressText.textContent = '文件更新成功';
                progressFill.style.backgroundColor = '#4CAF50';
                
                // 延迟后移除进度条并刷新列表
                setTimeout(() => {
                    progressContainer.remove();
                    loadFolderContents(currentFolderId);
                }, 2000);
            } else {
                throw new Error(result.msg || '更新文件失败');
            }
            
        } catch (error) {
            console.error('文件更新失败:', error);
            alert(`文件更新失败: ${error.message}`);
            
            // 移除进度条
            const progressContainer = document.querySelector('.upload-progress-container');
            if (progressContainer) {
                progressContainer.remove();
            }
        }
    };
    
    input.click();
}

// 获取文件版本
async function showFileVersions(fileId, fileName) {
    try {
        const response = await fetch(`${API_BASE}/file/versions/${fileId}`, {
            method: 'GET',
            credentials: 'include'
        });
        
        const result = await response.json();
        
        if (result.code === 20000) {
            const versions = result.data || [];
            
            // 创建版本列表模态框
            const modal = document.createElement('div');
            modal.className = 'versions-modal';
            modal.style.cssText = `
                position: fixed;
                top: 0;
                left: 0;
                width: 100%;
                height: 100%;
                background-color: rgba(0, 0, 0, 0.5);
                display: flex;
                justify-content: center;
                align-items: center;
                z-index: 1000;
            `;
            
            modal.innerHTML = `
                <div style="background: white; padding: 30px; border-radius: 8px; max-width: 600px; width: 90%; max-height: 80%; overflow-y: auto;">
                    <h2 style="margin-top: 0; margin-bottom: 20px; text-align: center;">文件版本历史</h2>
                    <h3 style="margin-bottom: 15px; color: #666;">${fileName}</h3>
                    ${versions.length > 0 ? `
                        <div style="line-height: 1.6;">
                            ${versions.map((version, index) => `
                                <div style="padding: 15px; margin-bottom: 10px; border: 1px solid #ddd; border-radius: 4px; ${index === 0 ? 'background-color: #f0f8ff;' : ''}">
                                    <div style="display: flex; justify-content: space-between; align-items: center;">
                                        <div>
                                            <strong>${version.name}</strong>
                                            ${index === 0 ? '<span style="color: #007bff; font-size: 12px; margin-left: 10px;">(当前版本)</span>' : ''}
                                        </div>
                                        <div style="font-size: 12px; color: #666;">
                                            ${formatFileSize(version.size)}
                                        </div>
                                    </div>
                                    <div style="font-size: 12px; color: #999; margin-top: 5px;">
                                        创建时间: ${formatDate(version.utime)}
                                    </div>
                                    <div style="font-size: 12px; color: #999;">
                                        哈希: ${version.hash}
                                    </div>
                                    ${version.deviceId ? `<div style="font-size: 12px; color: #999;">设备: ${version.deviceId}</div>` : ''}
                                </div>
                            `).join('')}
                        </div>
                    ` : '<p style="text-align: center; color: #666;">暂无版本信息</p>'}
                    <div style="text-align: center; margin-top: 20px;">
                        <button onclick="closeVersionsModal()" style="padding: 8px 20px; background: #007bff; color: white; border: none; border-radius: 4px; cursor: pointer;">关闭</button>
                    </div>
                </div>
            `;
            
            document.body.appendChild(modal);
            
            // 点击外部关闭
            modal.addEventListener('click', (e) => {
                if (e.target === modal) {
                    modal.remove();
                }
            });
            
        } else {
            alert(result.msg || '获取文件版本失败');
            if (result.code === 40001) {
                clearAuthInfo();
                showLoginPage();
            }
        }
    } catch (error) {
        console.error('获取文件版本失败:', error);
        alert('获取文件版本失败，请检查网络连接');
    }
}

// 关闭版本模态框
function closeVersionsModal() {
    const modal = document.querySelector('.versions-modal');
    if (modal) {
        modal.remove();
    }
}

// 批量操作功能
let selectedFiles = new Set();

// 切换文件选择状态
function toggleFileSelection(fileId, checkbox) {
    if (checkbox.checked) {
        selectedFiles.add(fileId);
    } else {
        selectedFiles.delete(fileId);
    }
    updateBatchToolbar();
}

// 全选/取消全选
function toggleSelectAll() {
    const checkboxes = document.querySelectorAll('.file-checkbox');
    const selectAllCheckbox = document.getElementById('selectAllCheckbox');
    const isSelectAll = selectAllCheckbox.checked;
    
    checkboxes.forEach(checkbox => {
        checkbox.checked = isSelectAll;
        const fileId = parseInt(checkbox.dataset.fileId);
        if (isSelectAll) {
            selectedFiles.add(fileId);
        } else {
            selectedFiles.delete(fileId);
        }
    });
    
    updateBatchToolbar();
}

// 更新批量操作工具栏
function updateBatchToolbar() {
    let toolbar = document.getElementById('batchToolbar');
    
    if (selectedFiles.size > 0) {
        // 显示批量操作工具栏
        if (!toolbar) {
            toolbar = document.createElement('div');
            toolbar.id = 'batchToolbar';
            toolbar.style.cssText = `
                position: fixed;
                bottom: 20px;
                left: 50%;
                transform: translateX(-50%);
                background: #fff;
                padding: 15px 20px;
                border-radius: 8px;
                box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
                border: 1px solid #ddd;
                z-index: 1000;
                display: flex;
                align-items: center;
                gap: 15px;
            `;
            
            toolbar.innerHTML = `
                <span style="color: #666; font-weight: 500;">已选择 <span id="selectedCount">${selectedFiles.size}</span> 个文件</span>
                <button id="batchDeleteBtn" class="btn-danger" style="padding: 8px 15px; background: #dc3545; color: white; border: none; border-radius: 4px; cursor: pointer;">批量删除</button>
                <button id="batchDownloadBtn" class="btn-secondary" style="padding: 8px 15px; background: #6c757d; color: white; border: none; border-radius: 4px; cursor: pointer;">批量下载</button>
                <button id="cancelBatchBtn" class="btn-light" style="padding: 8px 15px; background: #f8f9fa; color: #6c757d; border: 1px solid #ddd; border-radius: 4px; cursor: pointer;">取消选择</button>
            `;
            
            document.body.appendChild(toolbar);
            
            // 绑定事件
            document.getElementById('batchDeleteBtn').addEventListener('click', handleBatchDelete);
            document.getElementById('batchDownloadBtn').addEventListener('click', handleBatchDownload);
            document.getElementById('cancelBatchBtn').addEventListener('click', cancelBatchSelection);
        } else {
            // 更新选择数量
            document.getElementById('selectedCount').textContent = selectedFiles.size;
        }
    } else {
        // 隐藏批量操作工具栏
        if (toolbar) {
            toolbar.remove();
        }
    }
}

// 处理批量删除
async function handleBatchDelete() {
    const fileIds = Array.from(selectedFiles);
    await batchDeleteFiles(fileIds);
    cancelBatchSelection();
}

// 处理批量下载
async function handleBatchDownload() {
    const fileIds = Array.from(selectedFiles);
    alert(`批量下载功能开发中... 已选择 ${fileIds.length} 个文件`);
    // TODO: 实现批量下载功能
}

// 取消批量选择
function cancelBatchSelection() {
    selectedFiles.clear();
    
    // 取消所有复选框
    const checkboxes = document.querySelectorAll('.file-checkbox');
    checkboxes.forEach(checkbox => {
        checkbox.checked = false;
    });
    
    // 取消全选复选框
    const selectAllCheckbox = document.getElementById('selectAllCheckbox');
    if (selectAllCheckbox) {
        selectAllCheckbox.checked = false;
    }
    
    updateBatchToolbar();
}

// 设备版本管理功能
async function showDeviceVersionSelection(file) {
    try {
        // 获取用户的设备列表（假设有这个API）
        const devices = await getUserDevices();
        
        // 创建设备选择模态框
        const modal = document.createElement('div');
        modal.className = 'device-modal';
        modal.style.cssText = `
            position: fixed;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            background-color: rgba(0, 0, 0, 0.5);
            display: flex;
            justify-content: center;
            align-items: center;
            z-index: 1000;
        `;
        
        modal.innerHTML = `
            <div style="background: white; padding: 30px; border-radius: 8px; max-width: 500px; width: 90%;">
                <h2 style="margin-top: 0; margin-bottom: 20px; text-align: center;">选择设备版本</h2>
                <h3 style="margin-bottom: 15px; color: #666;">${file.name}</h3>
                <p style="color: #666; margin-bottom: 20px;">检测到多个设备上存在此文件，请选择要使用的版本：</p>
                <div style="max-height: 300px; overflow-y: auto; margin-bottom: 20px;">
                    ${devices.map(device => `
                        <div style="padding: 15px; margin-bottom: 10px; border: 1px solid #ddd; border-radius: 4px; cursor: pointer;" 
                             onclick="selectDeviceVersion('${device.id}', '${file.hash}')">
                            <div style="font-weight: bold;">${device.name}</div>
                            <div style="font-size: 12px; color: #666;">设备ID: ${device.id}</div>
                            <div style="font-size: 12px; color: #666;">最后同步: ${formatDate(device.lastSync)}</div>
                        </div>
                    `).join('')}
                </div>
                <div style="text-align: center;">
                    <button onclick="closeDeviceModal()" style="padding: 8px 20px; background: #6c757d; color: white; border: none; border-radius: 4px; cursor: pointer; margin-right: 10px;">取消</button>
                    <button onclick="createNewDeviceVersion('${file.id}')" style="padding: 8px 20px; background: #007bff; color: white; border: none; border-radius: 4px; cursor: pointer;">创建新版本</button>
                </div>
            </div>
        `;
        
        document.body.appendChild(modal);
        
        // 点击外部关闭
        modal.addEventListener('click', (e) => {
            if (e.target === modal) {
                modal.remove();
            }
        });
        
    } catch (error) {
        console.error('获取设备列表失败:', error);
        alert('获取设备列表失败');
    }
}

// 获取用户设备列表（模拟数据）
async function getUserDevices() {
    // TODO: 实现真实的设备API调用
    return [
        {
            id: 'device-001',
            name: '我的电脑',
            lastSync: Date.now() - 3600000 // 1小时前
        },
        {
            id: 'device-002', 
            name: '我的手机',
            lastSync: Date.now() - 7200000 // 2小时前
        },
        {
            id: 'device-003',
            name: '办公电脑',
            lastSync: Date.now() - 86400000 // 1天前
        }
    ];
}

// 选择设备版本
async function selectDeviceVersion(deviceId, fileHash) {
    try {
        // TODO: 实现选择特定设备版本的API调用
        alert(`已选择设备 ${deviceId} 的版本`);
        closeDeviceModal();
        loadFolderContents(currentFolderId);
    } catch (error) {
        console.error('选择设备版本失败:', error);
        alert('选择设备版本失败');
    }
}

// 创建新设备版本
async function createNewDeviceVersion(fileId) {
    try {
        // TODO: 实现创建新设备版本的API调用
        alert('创建新设备版本功能开发中...');
        closeDeviceModal();
    } catch (error) {
        console.error('创建新设备版本失败:', error);
        alert('创建新设备版本失败');
    }
}

// 关闭设备模态框
function closeDeviceModal() {
    const modal = document.querySelector('.device-modal');
    if (modal) {
        modal.remove();
    }
}

// 文件预览功能
function preWatchFile(fileUrl, fileName) {
    const fileExtension = fileName.split('.').pop().toLowerCase();
    
    // 创建预览模态框
    const modal = document.createElement('div');
    modal.className = 'preview-modal';
    modal.style.cssText = `
        position: fixed;
        top: 0;
        left: 0;
        width: 100%;
        height: 100%;
        background-color: rgba(0, 0, 0, 0.8);
        display: flex;
        justify-content: center;
        align-items: center;
        z-index: 1000;
    `;
    
    let previewContent = '';
    
    // 根据文件类型生成预览内容
    if (['jpg', 'jpeg', 'png', 'gif', 'bmp', 'webp'].includes(fileExtension)) {
        // 图片预览
        previewContent = `
            <div style="background: white; padding: 20px; border-radius: 8px; max-width: 90%; max-height: 90%; overflow: auto;">
                <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 15px;">
                    <h3 style="margin: 0; color: #333;">${fileName}</h3>
                    <button onclick="closePreviewModal()" style="background: none; border: none; font-size: 24px; cursor: pointer; color: #666;">&times;</button>
                </div>
                <img src="${fileUrl}" alt="${fileName}" style="max-width: 100%; max-height: 70vh; display: block; margin: 0 auto;">
                <div style="text-align: center; margin-top: 15px;">
                    <a href="${fileUrl}" download="${fileName}" style="padding: 8px 16px; background: #007bff; color: white; text-decoration: none; border-radius: 4px; margin-right: 10px;">下载</a>
                    <button onclick="copyFileUrl('${fileUrl}')" style="padding: 8px 16px; background: #28a745; color: white; border: none; border-radius: 4px;">复制链接</button>
                </div>
            </div>
        `;
    } else if (['mp4', 'avi', 'mov', 'wmv', 'flv', 'webm'].includes(fileExtension)) {
        // 视频预览
        previewContent = `
            <div style="background: white; padding: 20px; border-radius: 8px; max-width: 90%; max-height: 90%;">
                <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 15px;">
                    <h3 style="margin: 0; color: #333;">${fileName}</h3>
                    <button onclick="closePreviewModal()" style="background: none; border: none; font-size: 24px; cursor: pointer; color: #666;">&times;</button>
                </div>
                <video controls style="max-width: 100%; max-height: 70vh;">
                    <source src="${fileUrl}" type="video/${fileExtension}">
                    您的浏览器不支持视频播放。
                </video>
                <div style="text-align: center; margin-top: 15px;">
                    <a href="${fileUrl}" download="${fileName}" style="padding: 8px 16px; background: #007bff; color: white; text-decoration: none; border-radius: 4px; margin-right: 10px;">下载</a>
                    <button onclick="copyFileUrl('${fileUrl}')" style="padding: 8px 16px; background: #28a745; color: white; border: none; border-radius: 4px;">复制链接</button>
                </div>
            </div>
        `;
    } else if (['mp3', 'wav', 'ogg', 'aac', 'flac'].includes(fileExtension)) {
        // 音频预览
        previewContent = `
            <div style="background: white; padding: 20px; border-radius: 8px; max-width: 90%; max-height: 90%;">
                <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 15px;">
                    <h3 style="margin: 0; color: #333;">${fileName}</h3>
                    <button onclick="closePreviewModal()" style="background: none; border: none; font-size: 24px; cursor: pointer; color: #666;">&times;</button>
                </div>
                <div style="text-align: center; margin: 40px 0;">
                    <div style="font-size: 48px; color: #ccc; margin-bottom: 20px;">🎵</div>
                    <audio controls style="width: 100%; max-width: 400px;">
                        <source src="${fileUrl}" type="audio/${fileExtension}">
                        您的浏览器不支持音频播放。
                    </audio>
                </div>
                <div style="text-align: center; margin-top: 15px;">
                    <a href="${fileUrl}" download="${fileName}" style="padding: 8px 16px; background: #007bff; color: white; text-decoration: none; border-radius: 4px; margin-right: 10px;">下载</a>
                    <button onclick="copyFileUrl('${fileUrl}')" style="padding: 8px 16px; background: #28a745; color: white; border: none; border-radius: 4px;">复制链接</button>
                </div>
            </div>
        `;
    } else if (['pdf'].includes(fileExtension)) {
        // PDF预览
        previewContent = `
            <div style="background: white; padding: 20px; border-radius: 8px; max-width: 95%; max-height: 95%; display: flex; flex-direction: column;">
                <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 15px;">
                    <h3 style="margin: 0; color: #333;">${fileName}</h3>
                    <button onclick="closePreviewModal()" style="background: none; border: none; font-size: 24px; cursor: pointer; color: #666;">&times;</button>
                </div>
                <iframe src="${fileUrl}" style="flex: 1; width: 100%; min-height: 70vh; border: 1px solid #ddd;"></iframe>
                <div style="text-align: center; margin-top: 15px;">
                    <a href="${fileUrl}" download="${fileName}" style="padding: 8px 16px; background: #007bff; color: white; text-decoration: none; border-radius: 4px; margin-right: 10px;">下载</a>
                    <button onclick="copyFileUrl('${fileUrl}')" style="padding: 8px 16px; background: #28a745; color: white; border: none; border-radius: 4px;">复制链接</button>
                </div>
            </div>
        `;
    } else if (['txt', 'md', 'json', 'xml', 'html', 'css', 'js', 'py', 'java', 'cpp', 'c', 'go'].includes(fileExtension)) {
        // 文本文件预览
        previewContent = `
            <div style="background: white; padding: 20px; border-radius: 8px; max-width: 90%; max-height: 90%; display: flex; flex-direction: column;">
                <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 15px;">
                    <h3 style="margin: 0; color: #333;">${fileName}</h3>
                    <button onclick="closePreviewModal()" style="background: none; border: none; font-size: 24px; cursor: pointer; color: #666;">&times;</button>
                </div>
                <div style="flex: 1; min-height: 300px; border: 1px solid #ddd; padding: 15px; overflow: auto; background: #f8f9fa; font-family: monospace; font-size: 14px; line-height: 1.5;">
                    <div id="textPreviewContent">正在加载文件内容...</div>
                </div>
                <div style="text-align: center; margin-top: 15px;">
                    <a href="${fileUrl}" download="${fileName}" style="padding: 8px 16px; background: #007bff; color: white; text-decoration: none; border-radius: 4px; margin-right: 10px;">下载</a>
                    <button onclick="copyFileUrl('${fileUrl}')" style="padding: 8px 16px; background: #28a745; color: white; border: none; border-radius: 4px;">复制链接</button>
                </div>
            </div>
        `;
    } else {
        // 其他文件类型，显示文件信息
        previewContent = `
            <div style="background: white; padding: 30px; border-radius: 8px; max-width: 500px; width: 90%;">
                <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px;">
                    <h3 style="margin: 0; color: #333;">文件信息</h3>
                    <button onclick="closePreviewModal()" style="background: none; border: none; font-size: 24px; cursor: pointer; color: #666;">&times;</button>
                </div>
                <div style="text-align: center; margin: 40px 0;">
                    <div style="font-size: 48px; color: #ccc; margin-bottom: 20px;">📄</div>
                    <p style="font-weight: bold; margin-bottom: 10px;">${fileName}</p>
                    <p style="color: #666; margin-bottom: 20px;">此文件类型不支持在线预览</p>
                </div>
                <div style="text-align: center;">
                    <a href="${fileUrl}" download="${fileName}" style="padding: 12px 24px; background: #007bff; color: white; text-decoration: none; border-radius: 4px; margin-right: 10px;">下载文件</a>
                    <button onclick="copyFileUrl('${fileUrl}')" style="padding: 12px 24px; background: #28a745; color: white; border: none; border-radius: 4px;">复制链接</button>
                </div>
            </div>
        `;
    }
    
    modal.innerHTML = previewContent;
    document.body.appendChild(modal);
    
    // 如果是文本文件，加载内容
    if (['txt', 'md', 'json', 'xml', 'html', 'css', 'js', 'py', 'java', 'cpp', 'c', 'go'].includes(fileExtension)) {
        loadTextFileContent(fileUrl);
    }
    
    // 点击外部关闭
    modal.addEventListener('click', (e) => {
        if (e.target === modal) {
            closePreviewModal();
        }
    });
    
    // ESC键关闭
    document.addEventListener('keydown', function escHandler(e) {
        if (e.key === 'Escape') {
            closePreviewModal();
            document.removeEventListener('keydown', escHandler);
        }
    });
}

// 加载文本文件内容
async function loadTextFileContent(fileUrl) {
    try {
        const response = await fetch(fileUrl);
        const text = await response.text();
        const contentDiv = document.getElementById('textPreviewContent');
        if (contentDiv) {
            // 限制显示内容长度，避免大文件卡顿
            const maxLength = 10000; // 最多显示10KB内容
            let displayText = text;
            if (text.length > maxLength) {
                displayText = text.substring(0, maxLength) + '\n\n... (文件内容过长，仅显示前10KB)';
            }
            contentDiv.textContent = displayText;
        }
    } catch (error) {
        const contentDiv = document.getElementById('textPreviewContent');
        if (contentDiv) {
            contentDiv.textContent = '无法加载文件内容，可能是文件过大或网络错误。';
            contentDiv.style.color = '#dc3545';
        }
    }
}

// 关闭预览模态框
function closePreviewModal() {
    const modal = document.querySelector('.preview-modal');
    if (modal) {
        modal.remove();
    }
}

// 复制文件链接
function copyFileUrl(fileUrl) {
    navigator.clipboard.writeText(fileUrl).then(() => {
        // 创建临时提示
        const toast = document.createElement('div');
        toast.textContent = '链接已复制到剪贴板';
        toast.style.cssText = `
            position: fixed;
            top: 20px;
            right: 20px;
            background: #28a745;
            color: white;
            padding: 10px 20px;
            border-radius: 4px;
            z-index: 2000;
            animation: slideInRight 0.3s ease-out;
        `;
        
        document.body.appendChild(toast);
        
        setTimeout(() => {
            toast.remove();
        }, 2000);
    }).catch(err => {
        alert('复制失败，请手动复制链接');
        console.error('复制失败:', err);
    });
}
