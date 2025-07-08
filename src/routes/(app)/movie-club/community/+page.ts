import { error } from '@sveltejs/kit';
import axios from 'axios';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ fetch }) => {
	try {
		// Get user's accessible Matrix rooms
		const roomsResponse = await axios.get('/api/matrix/rooms');
		
		return {
			rooms: roomsResponse.data.rooms || []
		};
	} catch (err: any) {
		// If Matrix is not enabled or user has no access, return empty rooms
		if (err?.response?.status === 404 || err?.response?.status === 401) {
			return {
				rooms: []
			};
		}
		throw error(500, 'Failed to load community data');
	}
};