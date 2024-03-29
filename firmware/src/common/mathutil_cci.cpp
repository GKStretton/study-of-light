/* CircleCircleIntersection() *
 * Determine the points where 2 circles in a common plane intersect.
 *
 * int CircleCircleIntersection(
 *                                // center and radius of 1st circle
 *                                float x0, float y0, float r0,
 *                                // center and radius of 2nd circle
 *                                float x1, float y1, float r1,
 *                                // 1st intersection point
 *                                float *xi, float *yi,              
 *                                // 2nd intersection point
 *                                float *xi_prime, float *yi_prime)
 *
 * This is a public domain work. 3/26/2005 Tim Voght
 *
 */
#include "mathutil.h"
#include <stdio.h>
#include <math.h>

int CircleCircleIntersection(float x0, float y0, float r0,
															 float x1, float y1, float r1,
															 float *xi, float *yi,
															 float *xi_prime, float *yi_prime)
{
	float a, dx, dy, d, h, rx, ry;
	float x2, y2;

	/* dx and dy are the vertical and horizontal distances between
	 * the circle centers.
	 */
	dx = x1 - x0;
	dy = y1 - y0;

	/* Determine the straight-line distance between the centers. */
	//d = sqrt((dy*dy) + (dx*dx));
	d = hypotf(dx,dy); // Suggested by Keith Briggs

	/* Check for solvability. */
	if (d > (r0 + r1))
	{
		/* no solution. circles do not intersect. */
		return 0;
	}
	if (d < fabsf(r0 - r1))
	{
		/* no solution. one circle is contained in the other */
		return 0;
	}

	/* 'point 2' is the point where the line through the circle
	 * intersection points crosses the line between the circle
	 * centers.  
	 */

	/* Determine the distance from point 0 to point 2. */
	a = ((r0*r0) - (r1*r1) + (d*d)) / (2.0 * d) ;

	/* Determine the coordinates of point 2. */
	x2 = x0 + (dx * a/d);
	y2 = y0 + (dy * a/d);

	/* Determine the distance from point 2 to either of the
	 * intersection points.
	 */
	h = sqrtf((r0*r0) - (a*a));

	/* Now determine the offsets of the intersection points from
	 * point 2.
	 */
	rx = -dy * (h/d);
	ry = dx * (h/d);

	/* Determine the absolute intersection points. */
	*xi = x2 + rx;
	*xi_prime = x2 - rx;
	*yi = y2 + ry;
	*yi_prime = y2 - ry;

	return 1;
}