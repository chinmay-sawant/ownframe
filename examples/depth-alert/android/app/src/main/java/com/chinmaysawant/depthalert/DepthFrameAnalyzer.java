package com.chinmaysawant.depthalert;

import android.media.Image;

import java.nio.ByteBuffer;
import java.nio.ByteOrder;
import java.util.Arrays;

final class DepthFrameAnalyzer {
    static final class Summary {
        final int nearestMillimeters;
        final int coveragePercent;

        Summary(int nearestMillimeters, int coveragePercent) {
            this.nearestMillimeters = nearestMillimeters;
            this.coveragePercent = coveragePercent;
        }
    }

    static Summary center(Image depth) {
        Image.Plane depthPlane = depth.getPlanes()[0];
        ByteBuffer depths = depthPlane.getBuffer().order(ByteOrder.LITTLE_ENDIAN);
        int width = depth.getWidth();
        int height = depth.getHeight();
        int left = width * 3 / 10;
        int right = width * 7 / 10;
        int top = height * 3 / 10;
        int bottom = height * 7 / 10;
        int area = (right - left) * (bottom - top);
        int[] samples = new int[area];
        int[] neighborhood = new int[9];
        int count = 0;

        for (int y = top; y < bottom; y++) {
            for (int x = left; x < right; x++) {
                int nearby = 0;
                for (int dy = -1; dy <= 1; dy++) {
                    for (int dx = -1; dx <= 1; dx++) {
                        int px = x + dx;
                        int py = y + dy;
                        if (px < 0 || px >= width || py < 0 || py >= height) continue;
                        int offset = py * depthPlane.getRowStride()
                                + px * depthPlane.getPixelStride();
                        if (offset + 1 >= depths.limit()) continue;
                        int millimeters = Short.toUnsignedInt(depths.getShort(offset));
                        if (millimeters > 0 && millimeters <= 15000) {
                            neighborhood[nearby++] = millimeters;
                        }
                    }
                }
                if (nearby >= 5) {
                    Arrays.sort(neighborhood, 0, nearby);
                    samples[count++] = neighborhood[nearby / 2];
                }
            }
        }

        int coverage = area == 0 ? 0 : count * 100 / area;
        if (count == 0) {
            return new Summary(0, coverage);
        }
        Arrays.sort(samples, 0, count);
        int percentile = Math.min(count - 1, count / 10);
        return new Summary(samples[percentile], coverage);
    }

    private DepthFrameAnalyzer() {}
}
