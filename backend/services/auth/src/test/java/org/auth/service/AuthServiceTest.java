package org.auth.service;

import org.auth.dto.AuthResponse;
import org.auth.dto.LoginRequest;
import org.auth.dto.RegisterRequest;
import org.auth.dto.UserProfile;
import org.auth.exception.InvalidTokenException;
import org.auth.exception.ResourceNotFoundException;
import org.auth.exception.UserAlreadyExistsException;
import org.auth.model.Session;
import org.auth.model.User;
import org.auth.repository.SessionRepository;
import org.auth.repository.UserRepository;
import org.auth.util.JwtUtil;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.data.redis.core.ValueOperations;
import org.springframework.security.authentication.AuthenticationManager;
import org.springframework.security.authentication.UsernamePasswordAuthenticationToken;
import org.springframework.security.core.Authentication;
import org.springframework.security.core.context.SecurityContext;
import org.springframework.security.core.context.SecurityContextHolder;
import org.springframework.security.crypto.password.PasswordEncoder;

import java.time.LocalDateTime;
import java.util.Date;
import java.util.Optional;
import java.util.concurrent.TimeUnit;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.ArgumentMatchers.*;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class AuthServiceTest {

    @Mock
    private UserRepository userRepository;
    @Mock
    private SessionRepository sessionRepository;
    @Mock
    private PasswordEncoder passwordEncoder;
    @Mock
    private JwtUtil jwtUtil;
    @Mock
    private AuthenticationManager authenticationManager;
    @Mock
    private StringRedisTemplate redisTemplate;
    @Mock
    private ValueOperations<String, String> valueOperations;

    @InjectMocks
    private AuthService authService;

    private User user;
    private RegisterRequest registerRequest;
    private LoginRequest loginRequest;

    @BeforeEach
    void setUp() {
        user = User.builder()
                .id(1L)
                .email("test@example.com")
                .password("encodedPassword")
                .phone("+1234567890")
                .userType("CUSTOMER")
                .isActive(true)
                .build();

        registerRequest = new RegisterRequest("test@example.com", "password123", "+1234567890", "CUSTOMER");
        loginRequest = new LoginRequest("test@example.com", "password123", "iPhone");
    }

    @Test
    void register_Success() {
        when(userRepository.existsByEmail(anyString())).thenReturn(false);
        when(passwordEncoder.encode(anyString())).thenReturn("encodedPassword");
        when(jwtUtil.generateToken(anyString(), anyString(), any())).thenReturn("accessToken");
        when(jwtUtil.generateRefreshToken(anyString())).thenReturn("refreshToken");
        when(jwtUtil.extractExpiration(anyString())).thenReturn(new Date(System.currentTimeMillis() + 100000));

        AuthResponse response = authService.register(registerRequest);

        assertNotNull(response);
        assertEquals("accessToken", response.accessToken());
        assertEquals("refreshToken", response.refreshToken());
        verify(userRepository).save(any(User.class));
        verify(sessionRepository).save(any(Session.class));
    }

    @Test
    void register_UserAlreadyExists_ThrowsException() {
        when(userRepository.existsByEmail(anyString())).thenReturn(true);

        assertThrows(UserAlreadyExistsException.class, () -> authService.register(registerRequest));
    }

    @Test
    void login_Success() {
        when(userRepository.findByEmail(anyString())).thenReturn(Optional.of(user));
        when(jwtUtil.generateToken(anyString(), anyString(), any())).thenReturn("accessToken");
        when(jwtUtil.generateRefreshToken(anyString())).thenReturn("refreshToken");
        when(jwtUtil.extractExpiration(anyString())).thenReturn(new Date(System.currentTimeMillis() + 100000));

        AuthResponse response = authService.login(loginRequest);

        assertNotNull(response);
        assertEquals("accessToken", response.accessToken());
        assertEquals("refreshToken", response.refreshToken());
        verify(authenticationManager).authenticate(any(UsernamePasswordAuthenticationToken.class));
        verify(sessionRepository).save(any(Session.class));
    }

    @Test
    void login_UserNotFound_ThrowsException() {
        when(userRepository.findByEmail(anyString())).thenReturn(Optional.empty());

        assertThrows(ResourceNotFoundException.class, () -> authService.login(loginRequest));
    }

    @Test
    void logout_Success() {
        String authHeader = "Bearer validToken";
        Date expiration = new Date(System.currentTimeMillis() + 100000);
        when(jwtUtil.extractExpiration(anyString())).thenReturn(expiration);
        when(redisTemplate.opsForValue()).thenReturn(valueOperations);

        authService.logout(authHeader);

        verify(valueOperations).set(eq("blacklist:validToken"), eq("true"), anyLong(), eq(TimeUnit.MILLISECONDS));
    }

    @Test
    void refreshToken_Success() {
        Session session = Session.builder()
                .user(user)
                .token("refreshToken")
                .expiresAt(LocalDateTime.now().plusDays(1))
                .build();

        when(sessionRepository.findByToken(anyString())).thenReturn(Optional.of(session));
        when(jwtUtil.generateToken(anyString(), anyString(), any())).thenReturn("newAccessToken");

        AuthResponse response = authService.refreshToken("refreshToken");

        assertNotNull(response);
        assertEquals("newAccessToken", response.accessToken());
        assertEquals("refreshToken", response.refreshToken());
    }

    @Test
    void refreshToken_Expired_ThrowsException() {
        Session session = Session.builder()
                .user(user)
                .token("refreshToken")
                .expiresAt(LocalDateTime.now().minusDays(1))
                .build();

        when(sessionRepository.findByToken(anyString())).thenReturn(Optional.of(session));

        assertThrows(InvalidTokenException.class, () -> authService.refreshToken("refreshToken"));
        verify(sessionRepository).delete(session);
    }

    @Test
    void getCurrentUser_Success() {
        SecurityContext securityContext = mock(SecurityContext.class);
        Authentication authentication = mock(Authentication.class);
        when(securityContext.getAuthentication()).thenReturn(authentication);
        when(authentication.getName()).thenReturn("test@example.com");
        SecurityContextHolder.setContext(securityContext);

        when(userRepository.findByEmail("test@example.com")).thenReturn(Optional.of(user));

        UserProfile profile = authService.getCurrentUser();

        assertNotNull(profile);
        assertEquals(user.getEmail(), profile.email());
        assertEquals(user.getId(), profile.id());
    }
}
