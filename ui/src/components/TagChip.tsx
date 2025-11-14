import { memo } from 'react';
import { Chip } from '@mui/material';
import { LocalOffer } from '@mui/icons-material';

interface TagChipProps {
  tag: string;
  onDelete?: () => void;
  onClick?: () => void;
  size?: 'small' | 'medium';
  variant?: 'filled' | 'outlined';
}

function TagChip({ tag, onDelete, onClick, size = 'small', variant = 'filled' }: TagChipProps) {
  return (
    <Chip
      icon={<LocalOffer />}
      label={tag}
      size={size}
      variant={variant}
      onDelete={onDelete}
      onClick={onClick}
      color="primary"
      sx={{
        mr: 0.5,
        mb: 0.5,
        cursor: onClick ? 'pointer' : 'default'
      }}
    />
  );
}

// Memoize to prevent re-renders when tag hasn't changed
export default memo(TagChip);
