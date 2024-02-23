import { exec as execInternal } from 'child_process';
import { promisify } from 'util';

export const exec = promisify(execInternal)
