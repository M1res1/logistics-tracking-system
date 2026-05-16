package org.auth.service;

import lombok.RequiredArgsConstructor;
import org.auth.dto.AuthResponse;
import org.auth.dto.LoginRequest;
import org.auth.dto.RegisterRequest;
import org.auth.dto.UserProfile;
import org.auth.model.Session;
import org.auth.model.User;
import org.auth.repository.SessionRepository;
import org.auth.repository.UserRepository;
import org.auth.util.JwtUtil;
import org.auth.exception.InvalidTokenException;
import org.auth.exception.ResourceNotFoundException;
import org.auth.exception.UserAlreadyExistsException;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.security.authentication.AuthenticationManager;
import org.springframework.security.authentication.UsernamePasswordAuthenticationToken;
import org.springframework.security.core.context.SecurityContextHolder;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.LocalDateTime;
import java.time.ZoneId;
import java.util.Objects;
import java.util.concurrent.TimeUnit;

@Service
@RequiredArgsConstructor
public class AuthService {

    private final UserRepository userRepository;
    private final SessionRepository sessionRepository;
    private final PasswordEncoder passwordEncoder;
    private final JwtUtil jwtUtil;
    private final AuthenticationManager authenticationManager;
    private final StringRedisTemplate redisTemplate;

    @Transactional
    public AuthResponse register(RegisterRequest request) {
        if (userRepository.existsByEmail(request.email())) {
            throw new UserAlreadyExistsException("Email already exists");
        }

        User user = User.builder()
                .email(request.email())
                .password(passwordEncoder.encode(request.password()))
                .phone(request.phone())
                .userType(request.userType())
                .isActive(true)
                .build();

        userRepository.save(user);

        String accessToken = jwtUtil.generateToken(user.getEmail(), user.getUserType(), user.getId());
        String refreshToken = jwtUtil.generateRefreshToken(user.getEmail());

        saveSession(user, refreshToken, null);

        return AuthResponse.builder()
                .accessToken(accessToken)
                .refreshToken(refreshToken)
                .build();
    }

    @Transactional
    public AuthResponse login(LoginRequest request) {
        authenticationManager.authenticate(
                new UsernamePasswordAuthenticationToken(request.email(), request.password())
        );

        User user = userRepository.findByEmail(request.email())
                .orElseThrow(() -> new ResourceNotFoundException("User not found"));

        String accessToken = jwtUtil.generateToken(user.getEmail(), user.getUserType(), user.getId());
        String refreshToken = jwtUtil.generateRefreshToken(user.getEmail());

        saveSession(user, refreshToken, request.deviceInfo());

        return AuthResponse.builder()
                .accessToken(accessToken)
                .refreshToken(refreshToken)
                .build();
    }

    @Transactional
    public void logout(String authHeader) {
        if (authHeader != null && authHeader.startsWith("Bearer ")) {
            String token = authHeader.substring(7);
            long expiry = jwtUtil.extractExpiration(token).getTime() - System.currentTimeMillis();
            if (expiry > 0) {
                redisTemplate.opsForValue().set("blacklist:" + token, "true", expiry, TimeUnit.MILLISECONDS);
            }
        }
    }

    @Transactional
    public AuthResponse refreshToken(String refreshToken) {
        Session session = sessionRepository.findByToken(refreshToken)
                .orElseThrow(() -> new InvalidTokenException("Invalid refresh token"));

        if (session.getExpiresAt().isBefore(LocalDateTime.now())) {
            sessionRepository.delete(session);
            throw new InvalidTokenException("Refresh token expired");
        }

        User user = session.getUser();
        String newAccessToken = jwtUtil.generateToken(user.getEmail(), user.getUserType(), user.getId());

        return AuthResponse.builder()
                .accessToken(newAccessToken)
                .refreshToken(refreshToken)
                .build();
    }

    public UserProfile getCurrentUser() {
        String email = Objects.requireNonNull(SecurityContextHolder.getContext().getAuthentication()).getName();
        User user = userRepository.findByEmail(email)
                .orElseThrow(() -> new ResourceNotFoundException("User not found"));

        return UserProfile.builder()
                .id(user.getId())
                .email(user.getEmail())
                .phone(user.getPhone())
                .userType(user.getUserType())
                .build();
    }

    private void saveSession(User user, String token, String deviceInfo) {
        LocalDateTime expiresAt = jwtUtil.extractExpiration(token)
                .toInstant().atZone(ZoneId.systemDefault()).toLocalDateTime();

        Session session = Session.builder()
                .user(user)
                .token(token)
                .expiresAt(expiresAt)
                .deviceInfo(deviceInfo)
                .build();

        sessionRepository.save(session);
    }
}
