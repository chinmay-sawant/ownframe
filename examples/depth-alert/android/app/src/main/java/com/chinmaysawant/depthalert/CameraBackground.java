package com.chinmaysawant.depthalert;

import android.opengl.GLES11Ext;
import android.opengl.GLES20;

import com.google.ar.core.Coordinates2d;
import com.google.ar.core.Frame;

import java.nio.ByteBuffer;
import java.nio.ByteOrder;
import java.nio.FloatBuffer;

final class CameraBackground {
    private static final float[] QUAD = {-1, -1, 1, -1, -1, 1, 1, 1};
    private static final String VERTEX =
            "attribute vec2 aPosition; attribute vec2 aUv; varying vec2 vUv; "
                    + "void main(){ gl_Position=vec4(aPosition,0.0,1.0); vUv=aUv; }";
    private static final String FRAGMENT =
            "#extension GL_OES_EGL_image_external : require\n"
                    + "precision mediump float; uniform samplerExternalOES sTexture; "
                    + "varying vec2 vUv; void main(){ gl_FragColor=texture2D(sTexture,vUv); }";

    private final FloatBuffer vertices = floats(QUAD);
    private final FloatBuffer uvs = floats(new float[8]);
    private int program;
    private int texture;
    private int positionLocation;
    private int uvLocation;
    private int samplerLocation;

    int create() {
        program = link(shader(GLES20.GL_VERTEX_SHADER, VERTEX),
                shader(GLES20.GL_FRAGMENT_SHADER, FRAGMENT));
        positionLocation = GLES20.glGetAttribLocation(program, "aPosition");
        uvLocation = GLES20.glGetAttribLocation(program, "aUv");
        samplerLocation = GLES20.glGetUniformLocation(program, "sTexture");
        int[] textures = new int[1];
        GLES20.glGenTextures(1, textures, 0);
        texture = textures[0];
        GLES20.glBindTexture(GLES11Ext.GL_TEXTURE_EXTERNAL_OES, texture);
        GLES20.glTexParameteri(GLES11Ext.GL_TEXTURE_EXTERNAL_OES,
                GLES20.GL_TEXTURE_MIN_FILTER, GLES20.GL_LINEAR);
        GLES20.glTexParameteri(GLES11Ext.GL_TEXTURE_EXTERNAL_OES,
                GLES20.GL_TEXTURE_MAG_FILTER, GLES20.GL_LINEAR);
        GLES20.glTexParameteri(GLES11Ext.GL_TEXTURE_EXTERNAL_OES,
                GLES20.GL_TEXTURE_WRAP_S, GLES20.GL_CLAMP_TO_EDGE);
        GLES20.glTexParameteri(GLES11Ext.GL_TEXTURE_EXTERNAL_OES,
                GLES20.GL_TEXTURE_WRAP_T, GLES20.GL_CLAMP_TO_EDGE);
        return texture;
    }

    void draw(Frame frame) {
        if (frame.getTimestamp() == 0) return;
        float[] transformed = new float[8];
        frame.transformCoordinates2d(Coordinates2d.OPENGL_NORMALIZED_DEVICE_COORDINATES,
                QUAD, Coordinates2d.TEXTURE_NORMALIZED, transformed);
        uvs.clear();
        uvs.put(transformed).position(0);
        GLES20.glDisable(GLES20.GL_DEPTH_TEST);
        GLES20.glUseProgram(program);
        GLES20.glActiveTexture(GLES20.GL_TEXTURE0);
        GLES20.glBindTexture(GLES11Ext.GL_TEXTURE_EXTERNAL_OES, texture);
        GLES20.glUniform1i(samplerLocation, 0);
        GLES20.glEnableVertexAttribArray(positionLocation);
        GLES20.glEnableVertexAttribArray(uvLocation);
        GLES20.glVertexAttribPointer(positionLocation, 2, GLES20.GL_FLOAT,
                false, 0, vertices);
        GLES20.glVertexAttribPointer(uvLocation, 2, GLES20.GL_FLOAT,
                false, 0, uvs);
        GLES20.glDrawArrays(GLES20.GL_TRIANGLE_STRIP, 0, 4);
        GLES20.glDisableVertexAttribArray(positionLocation);
        GLES20.glDisableVertexAttribArray(uvLocation);
    }

    private static FloatBuffer floats(float[] values) {
        FloatBuffer result = ByteBuffer.allocateDirect(values.length * 4)
                .order(ByteOrder.nativeOrder()).asFloatBuffer();
        result.put(values).position(0);
        return result;
    }

    private static int shader(int type, String source) {
        int shader = GLES20.glCreateShader(type);
        GLES20.glShaderSource(shader, source);
        GLES20.glCompileShader(shader);
        int[] status = new int[1];
        GLES20.glGetShaderiv(shader, GLES20.GL_COMPILE_STATUS, status, 0);
        if (status[0] == 0) throw new IllegalStateException(GLES20.glGetShaderInfoLog(shader));
        return shader;
    }

    private static int link(int vertex, int fragment) {
        int program = GLES20.glCreateProgram();
        GLES20.glAttachShader(program, vertex);
        GLES20.glAttachShader(program, fragment);
        GLES20.glLinkProgram(program);
        int[] status = new int[1];
        GLES20.glGetProgramiv(program, GLES20.GL_LINK_STATUS, status, 0);
        if (status[0] == 0) throw new IllegalStateException(GLES20.glGetProgramInfoLog(program));
        return program;
    }
}
