import { useState } from 'react';
import {
  Box,
  Button,
  Card,
  CardContent,
  Container,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  IconButton,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  TextField,
  Typography,
  Chip,
  Alert,
  CircularProgress,
  Paper,
  InputAdornment,
  MenuItem,
  Select,
  FormControl,
  InputLabel,
  Grid,
  Tooltip,
  Stack,
  List,
  ListItem,
  ListItemText,
} from '@mui/material';
import {
  Add as AddIcon,
  Refresh as RefreshIcon,
  Delete as DeleteIcon,
  Edit as EditIcon,
  Search as SearchIcon,
  CheckCircle as OnlineIcon,
  Cancel as OfflineIcon,
  Help as UnknownIcon,
  Cloud as K8sIcon,
  Storage as DockerIcon,
} from '@mui/icons-material';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useNotificationQueue } from '../../components/NotificationQueue';
import AdminPortalLayout from '../../components/AdminPortalLayout';

/**
 * Controllers - Platform controller management
 *
 * Administrative interface for managing distributed platform controllers
 * that handle workloads across different infrastructure platforms.
 *
 * Features:
 * - List all registered controllers
 * - Register new controllers (K8s, Docker, Hyper-V, etc.)
 * - Update controller metadata
 * - Unregister controllers
 * - Monitor controller status and heartbeat
 * - View controller capabilities
 *
 * Controller Platforms:
 * - kubernetes: Kubernetes cluster controller
 * - docker: Docker host controller
 * - hyperv: Hyper-V controller
 * - vcenter: VMware vCenter controller
 *
 * Controller Status:
 * - connected: Controller is online and responding
 * - disconnected: Controller is offline
 * - unknown: Status unknown
 *
 * @page
 * @route /admin/controllers - Controller management
 * @access admin - Restricted to administrators only
 *
 * @component
 *
 * @returns {JSX.Element} Controller management interface
 */
export default function Controllers() {
  const { addNotification } = useNotificationQueue();
  const queryClient = useQueryClient();

  const [searchQuery, setSearchQuery] = useState('');
  const [platformFilter, setPlatformFilter] = useState<string>('all');
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [registerDialogOpen, setRegisterDialogOpen] = useState(false);
  const [editDialogOpen, setEditDialogOpen] = useState(false);
  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false);
  const [selectedController, setSelectedController] = useState<any>(null);

  // Form state
  const [formData, setFormData] = useState({
    controller_id: '',
    platform: 'kubernetes',
    display_name: '',
    version: '',
    capabilities: [] as string[],
  });

  // Available platforms
  const platforms = [
    { value: 'kubernetes', label: 'Kubernetes', icon: <K8sIcon /> },
    { value: 'docker', label: 'Docker', icon: <DockerIcon /> },
    { value: 'hyperv', label: 'Hyper-V' },
    { value: 'vcenter', label: 'VMware vCenter' },
  ];

  // Fetch controllers
  const { data: controllers, isLoading, refetch } = useQuery({
    queryKey: ['controllers', platformFilter, statusFilter],
    queryFn: async () => {
      const params = new URLSearchParams();
      if (platformFilter !== 'all') {
        params.append('platform', platformFilter);
      }
      if (statusFilter !== 'all') {
        params.append('status', statusFilter);
      }

      const response = await fetch(`/api/v1/admin/controllers?${params}`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
      });

      if (!response.ok) {
        throw new Error('Failed to fetch controllers');
      }

      return response.json();
    },
  });

  // Register controller mutation
  const registerMutation = useMutation({
    mutationFn: async (data: typeof formData) => {
      const response = await fetch('/api/v1/admin/controllers/register', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(data),
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.message || 'Failed to register controller');
      }

      return response.json();
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['controllers'] });
      setRegisterDialogOpen(false);
      setFormData({
        controller_id: '',
        platform: 'kubernetes',
        display_name: '',
        version: '',
        capabilities: [],
      });
      addNotification({
        message: 'Controller registered successfully',
        severity: 'success',
        priority: 'high',
        title: 'Controller Registered',
      });
    },
    onError: (error: Error) => {
      addNotification({
        message: `Failed to register controller: ${error.message}`,
        severity: 'error',
        priority: 'high',
        title: 'Registration Failed',
      });
    },
  });

  // Update controller mutation
  const updateMutation = useMutation({
    mutationFn: async ({ id, data }: { id: string; data: Partial<typeof formData> }) => {
      const response = await fetch(`/api/v1/admin/controllers/${id}`, {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(data),
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.message || 'Failed to update controller');
      }

      return response.json();
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['controllers'] });
      setEditDialogOpen(false);
      setSelectedController(null);
      addNotification({
        message: 'Controller updated successfully',
        severity: 'success',
        priority: 'medium',
        title: 'Controller Updated',
      });
    },
    onError: (error: Error) => {
      addNotification({
        message: `Failed to update controller: ${error.message}`,
        severity: 'error',
        priority: 'high',
        title: 'Update Failed',
      });
    },
  });

  // Delete controller mutation
  const deleteMutation = useMutation({
    mutationFn: async (id: string) => {
      const response = await fetch(`/api/v1/admin/controllers/${id}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.message || 'Failed to delete controller');
      }

      return response.json();
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['controllers'] });
      setDeleteConfirmOpen(false);
      setSelectedController(null);
      addNotification({
        message: 'Controller unregistered successfully',
        severity: 'success',
        priority: 'medium',
        title: 'Controller Unregistered',
      });
    },
    onError: (error: Error) => {
      addNotification({
        message: `Failed to unregister controller: ${error.message}`,
        severity: 'error',
        priority: 'high',
        title: 'Delete Failed',
      });
    },
  });

  const handleRegisterController = () => {
    registerMutation.mutate(formData);
  };

  const handleUpdateController = () => {
    if (selectedController) {
      updateMutation.mutate({ id: selectedController.id, data: formData });
    }
  };

  const handleEditClick = (controller: any) => {
    setSelectedController(controller);
    setFormData({
      controller_id: controller.controller_id,
      platform: controller.platform,
      display_name: controller.display_name,
      version: controller.version,
      capabilities: controller.capabilities || [],
    });
    setEditDialogOpen(true);
  };

  const handleDeleteClick = (controller: any) => {
    setSelectedController(controller);
    setDeleteConfirmOpen(true);
  };

  const getStatusColor = (status: string) => {
    switch (status?.toLowerCase()) {
      case 'connected':
        return 'success';
      case 'disconnected':
        return 'error';
      case 'unknown':
      default:
        return 'default';
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status?.toLowerCase()) {
      case 'connected':
        return <OnlineIcon />;
      case 'disconnected':
        return <OfflineIcon />;
      case 'unknown':
      default:
        return <UnknownIcon />;
    }
  };

  const getPlatformIcon = (platform: string) => {
    const platformConfig = platforms.find(p => p.value === platform);
    return platformConfig?.icon || null;
  };

  const filteredControllers = (controllers || []).filter((ctrl: any) => {
    return (
      ctrl.display_name?.toLowerCase().includes(searchQuery.toLowerCase()) ||
      ctrl.controller_id?.toLowerCase().includes(searchQuery.toLowerCase()) ||
      ctrl.platform?.toLowerCase().includes(searchQuery.toLowerCase())
    );
  });

  const connectedCount = (controllers || []).filter((c: any) => c.status === 'connected').length;
  const disconnectedCount = (controllers || []).filter((c: any) => c.status === 'disconnected').length;

  if (isLoading) {
    return (
      <AdminPortalLayout title="Controllers">
        <Container maxWidth="xl">
          <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: 400 }}>
            <CircularProgress />
          </Box>
        </Container>
      </AdminPortalLayout>
    );
  }

  return (
    <AdminPortalLayout title="Platform Controllers">
      <Container maxWidth="xl">
        {/* Header */}
        <Box sx={{ mb: 3, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <Box>
            <Typography variant="h4" gutterBottom>
              Platform Controllers
            </Typography>
            <Typography variant="body2" color="text.secondary">
              Manage distributed controllers for multi-platform workloads
            </Typography>
          </Box>
          <Box sx={{ display: 'flex', gap: 1 }}>
            <Button
              variant="outlined"
              startIcon={<RefreshIcon />}
              onClick={() => refetch()}
            >
              Refresh
            </Button>
            <Button
              variant="contained"
              startIcon={<AddIcon />}
              onClick={() => setRegisterDialogOpen(true)}
            >
              Register Controller
            </Button>
          </Box>
        </Box>

        {/* Status Summary Cards */}
        <Grid container spacing={2} sx={{ mb: 3 }}>
          <Grid item xs={12} md={4}>
            <Card>
              <CardContent>
                <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                  <Box>
                    <Typography variant="h3">
                      {controllers?.length || 0}
                    </Typography>
                    <Typography variant="body2" color="text.secondary">
                      Total Controllers
                    </Typography>
                  </Box>
                  <Cloud sx={{ fontSize: 48, opacity: 0.3 }} />
                </Box>
              </CardContent>
            </Card>
          </Grid>
          <Grid item xs={12} md={4}>
            <Card>
              <CardContent>
                <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                  <Box>
                    <Typography variant="h3" color="success.main">
                      {connectedCount}
                    </Typography>
                    <Typography variant="body2" color="text.secondary">
                      Connected
                    </Typography>
                  </Box>
                  <OnlineIcon sx={{ fontSize: 48, color: 'success.main', opacity: 0.3 }} />
                </Box>
              </CardContent>
            </Card>
          </Grid>
          <Grid item xs={12} md={4}>
            <Card>
              <CardContent>
                <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                  <Box>
                    <Typography variant="h3" color="error.main">
                      {disconnectedCount}
                    </Typography>
                    <Typography variant="body2" color="text.secondary">
                      Disconnected
                    </Typography>
                  </Box>
                  <OfflineIcon sx={{ fontSize: 48, color: 'error.main', opacity: 0.3 }} />
                </Box>
              </CardContent>
            </Card>
          </Grid>
        </Grid>

        {/* Filters */}
        <Card sx={{ mb: 3 }}>
          <CardContent>
            <Grid container spacing={2} alignItems="center">
              <Grid item xs={12} md={4}>
                <TextField
                  fullWidth
                  placeholder="Search controllers..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  InputProps={{
                    startAdornment: (
                      <InputAdornment position="start">
                        <SearchIcon />
                      </InputAdornment>
                    ),
                  }}
                />
              </Grid>
              <Grid item xs={12} md={3}>
                <FormControl fullWidth>
                  <InputLabel>Platform</InputLabel>
                  <Select
                    value={platformFilter}
                    label="Platform"
                    onChange={(e) => setPlatformFilter(e.target.value)}
                  >
                    <MenuItem value="all">All Platforms</MenuItem>
                    {platforms.map((p) => (
                      <MenuItem key={p.value} value={p.value}>{p.label}</MenuItem>
                    ))}
                  </Select>
                </FormControl>
              </Grid>
              <Grid item xs={12} md={3}>
                <FormControl fullWidth>
                  <InputLabel>Status</InputLabel>
                  <Select
                    value={statusFilter}
                    label="Status"
                    onChange={(e) => setStatusFilter(e.target.value)}
                  >
                    <MenuItem value="all">All Status</MenuItem>
                    <MenuItem value="connected">Connected</MenuItem>
                    <MenuItem value="disconnected">Disconnected</MenuItem>
                    <MenuItem value="unknown">Unknown</MenuItem>
                  </Select>
                </FormControl>
              </Grid>
              <Grid item xs={12} md={2}>
                <Typography variant="body2" color="text.secondary">
                  Showing {filteredControllers.length} controllers
                </Typography>
              </Grid>
            </Grid>
          </CardContent>
        </Card>

        {/* Controllers Table */}
        <TableContainer component={Paper}>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell>Controller</TableCell>
                <TableCell>Platform</TableCell>
                <TableCell>Status</TableCell>
                <TableCell>Version</TableCell>
                <TableCell>Capabilities</TableCell>
                <TableCell>Last Heartbeat</TableCell>
                <TableCell>Actions</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {filteredControllers.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={7} align="center">
                    <Typography variant="body2" color="text.secondary">
                      No controllers found
                    </Typography>
                  </TableCell>
                </TableRow>
              ) : (
                filteredControllers.map((ctrl: any) => (
                  <TableRow key={ctrl.id}>
                    <TableCell>
                      <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                        {getPlatformIcon(ctrl.platform)}
                        <Box>
                          <Typography variant="body2" fontWeight="medium">
                            {ctrl.display_name}
                          </Typography>
                          <Typography variant="caption" color="text.secondary" sx={{ fontFamily: 'monospace' }}>
                            {ctrl.controller_id}
                          </Typography>
                        </Box>
                      </Box>
                    </TableCell>
                    <TableCell>
                      <Chip label={ctrl.platform} size="small" />
                    </TableCell>
                    <TableCell>
                      <Chip
                        icon={getStatusIcon(ctrl.status)}
                        label={ctrl.status}
                        color={getStatusColor(ctrl.status)}
                        size="small"
                      />
                    </TableCell>
                    <TableCell>
                      <Typography variant="body2">
                        {ctrl.version || 'N/A'}
                      </Typography>
                    </TableCell>
                    <TableCell>
                      <Stack direction="row" spacing={0.5} flexWrap="wrap">
                        {(ctrl.capabilities || []).slice(0, 2).map((cap: string) => (
                          <Chip key={cap} label={cap} size="small" variant="outlined" />
                        ))}
                        {(ctrl.capabilities || []).length > 2 && (
                          <Chip label={`+${ctrl.capabilities.length - 2}`} size="small" variant="outlined" />
                        )}
                      </Stack>
                    </TableCell>
                    <TableCell>
                      {ctrl.last_heartbeat ? (
                        <Typography variant="body2">
                          {new Date(ctrl.last_heartbeat).toLocaleString()}
                        </Typography>
                      ) : (
                        <Typography variant="body2" color="text.secondary">
                          Never
                        </Typography>
                      )}
                    </TableCell>
                    <TableCell>
                      <Stack direction="row" spacing={0.5}>
                        <Tooltip title="Edit">
                          <IconButton
                            size="small"
                            onClick={() => handleEditClick(ctrl)}
                          >
                            <EditIcon fontSize="small" />
                          </IconButton>
                        </Tooltip>
                        <Tooltip title="Unregister">
                          <IconButton
                            size="small"
                            color="error"
                            onClick={() => handleDeleteClick(ctrl)}
                          >
                            <DeleteIcon fontSize="small" />
                          </IconButton>
                        </Tooltip>
                      </Stack>
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </TableContainer>

        {/* Register Controller Dialog */}
        <Dialog
          open={registerDialogOpen}
          onClose={() => setRegisterDialogOpen(false)}
          maxWidth="sm"
          fullWidth
        >
          <DialogTitle>Register Platform Controller</DialogTitle>
          <DialogContent>
            <Typography variant="body2" color="text.secondary" paragraph>
              Register a new platform controller to distribute workloads.
            </Typography>
            <TextField
              fullWidth
              label="Controller ID"
              value={formData.controller_id}
              onChange={(e) => setFormData({ ...formData, controller_id: e.target.value })}
              sx={{ mt: 2, mb: 2 }}
              required
              helperText="Unique identifier for this controller"
            />
            <FormControl fullWidth sx={{ mb: 2 }}>
              <InputLabel>Platform</InputLabel>
              <Select
                value={formData.platform}
                label="Platform"
                onChange={(e) => setFormData({ ...formData, platform: e.target.value })}
              >
                {platforms.map((p) => (
                  <MenuItem key={p.value} value={p.value}>{p.label}</MenuItem>
                ))}
              </Select>
            </FormControl>
            <TextField
              fullWidth
              label="Display Name"
              value={formData.display_name}
              onChange={(e) => setFormData({ ...formData, display_name: e.target.value })}
              sx={{ mb: 2 }}
              helperText="Human-readable name for this controller"
            />
            <TextField
              fullWidth
              label="Version"
              value={formData.version}
              onChange={(e) => setFormData({ ...formData, version: e.target.value })}
              helperText="Controller version (e.g., v1.0.0)"
            />
          </DialogContent>
          <DialogActions>
            <Button onClick={() => setRegisterDialogOpen(false)}>
              Cancel
            </Button>
            <Button
              onClick={handleRegisterController}
              variant="contained"
              disabled={!formData.controller_id || !formData.platform || registerMutation.isPending}
            >
              {registerMutation.isPending ? 'Registering...' : 'Register'}
            </Button>
          </DialogActions>
        </Dialog>

        {/* Edit Controller Dialog */}
        <Dialog
          open={editDialogOpen}
          onClose={() => {
            setEditDialogOpen(false);
            setSelectedController(null);
          }}
          maxWidth="sm"
          fullWidth
        >
          <DialogTitle>Edit Controller</DialogTitle>
          <DialogContent>
            <TextField
              fullWidth
              label="Display Name"
              value={formData.display_name}
              onChange={(e) => setFormData({ ...formData, display_name: e.target.value })}
              sx={{ mt: 2, mb: 2 }}
            />
            <TextField
              fullWidth
              label="Version"
              value={formData.version}
              onChange={(e) => setFormData({ ...formData, version: e.target.value })}
            />
          </DialogContent>
          <DialogActions>
            <Button onClick={() => {
              setEditDialogOpen(false);
              setSelectedController(null);
            }}>
              Cancel
            </Button>
            <Button
              onClick={handleUpdateController}
              variant="contained"
              disabled={updateMutation.isPending}
            >
              {updateMutation.isPending ? 'Updating...' : 'Update'}
            </Button>
          </DialogActions>
        </Dialog>

        {/* Delete Confirmation Dialog */}
        <Dialog
          open={deleteConfirmOpen}
          onClose={() => {
            setDeleteConfirmOpen(false);
            setSelectedController(null);
          }}
          maxWidth="xs"
        >
          <DialogTitle>Unregister Controller?</DialogTitle>
          <DialogContent>
            <Typography>
              Are you sure you want to unregister this controller? Workloads may be affected.
            </Typography>
          </DialogContent>
          <DialogActions>
            <Button onClick={() => {
              setDeleteConfirmOpen(false);
              setSelectedController(null);
            }}>
              Cancel
            </Button>
            <Button
              onClick={() => selectedController && deleteMutation.mutate(selectedController.id)}
              color="error"
              variant="contained"
              disabled={deleteMutation.isPending}
            >
              {deleteMutation.isPending ? 'Unregistering...' : 'Unregister'}
            </Button>
          </DialogActions>
        </Dialog>
      </Container>
    </AdminPortalLayout>
  );
}
