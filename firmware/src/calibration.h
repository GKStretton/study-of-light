#pragma once

// Basic Stepper calibration

// After a limit switch is pressed, the current unit is set to the min unit minus
// this value. So the motor would have to move at least this many units below
// minimum to re-trigger the limit switch
#define CALIBRATION_GAP_SIZE 1

// multiplier for stepper speeds and accelerations. Useful for cautious testing.
#define SPEED_MULT 1

// ul of air to draw in from 0 position before taking in liquid
#define PIPETTE_BUFFER 50
// ul of intake during rinse
#define PIPETTE_RINSE 100
// z level from which to take in liquid. (may become dynamic based on dye level in future.)
#define PIPETTE_INTAKE_Z 40
// The amount of backlash / slop in ul when changing from collect to dispense.
#define PIPETTE_BACKLASH_UL 3

// Physical calibration of arm positions etc
#define STAGE_RADIUS_MM 75
#define ARM_PATH_RADIUS_MM 164.7f
// The stage is equivalent to the crop area. But the tip can't target -1 to 1.
// IK target requests will be bounded to a circle THIS portion of the stage radius
#define IK_TARGET_RADIUS_FRAC 0.8f

#define YAW_ZERO_OFFSET -21.3f
// anticlockwise offset of arm from negative x axis
#define RING_ZERO_OFFSET 4.8f

#define CENTRE_PITCH 48.95f
#define MIN_BOWL_Z 32.5f

// NODE CALIBRATION (UNITS)
#define HOME_TOP_Z 73

#define VIAL_PITCH 0
#define VIAL_YAW_OFFSET -6
#define VIAL_YAW_INCREMENT 36.38f

// how much to drop the z level during dispense, after pipette actuation.
#define DISPENSE_Z_OFFSET -4.0f

#define HANDOVER_Z 50
#define HANDOVER_PITCH 70
#define HANDOVER_INNER_YAW 15
#define HANDOVER_OUTER_YAW 50

// the maximum value of yaw that it's safe to enter ik mode within
#define MAX_BOWL_YAW -YAW_ZERO_OFFSET

// number of milliseconds to open the air valve for after fluid dispense
#define FLUID_TRAVEL_TIME_MS 5000
#define OPEN_DRAIN_DELAY_MS 8000
#define WATER_VOLUME_PER_SECOND_ML 8.25f
#define MILK_VOLUME_PER_SECOND_ML 7.5f
#define DRAIN_VOLUME_PER_SECOND_ML 4.5f
#define MAX_FLUID_LEVEL 500.0f

// Rinse node calibration
#define RINSE_CONTAINER_ENTRY_Z 67
#define RINSE_CONTAINER_LOW_Z 30
#define RINSE_CONTAINER_PITCH 11
#define RINSE_CONTAINER_YAW 63

#define MAINTENANCE_RING_ANGLE 110
#define IDLE_RING_ANGLE 16

// Finer calibration for IK movements. See Z variance calibration google sheet.
// z mm offset per yaw degree. 
#define IK_Z_CALIB_YAW_MM_PER_DEG 0.08111f