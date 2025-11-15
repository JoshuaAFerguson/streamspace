import { useState } from 'react';
import {
  Box,
  Typography,
  Tabs,
  Tab,
  Card,
  CardContent,
  Button,
  Chip,
  IconButton,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  Select,
  MenuItem,
  FormControl,
  InputLabel,
  Switch,
  FormControlLabel,
  Alert,
  Grid,
} from '@mui/material';
import {
  Webhook as WebhookIcon,
  Add as AddIcon,
  Edit as EditIcon,
  Delete as DeleteIcon,
  PlayArrow as TestIcon,
  History as HistoryIcon,
  CheckCircle as SuccessIcon,
  Error as ErrorIcon,
  Pending as PendingIcon,
} from '@mui/icons-material';
import Layout from '../../components/Layout';

interface Webhook {
  id: number;
  name: string;
  url: string;
  events: string[];
  enabled: boolean;
  created_at: string;
}

interface WebhookDelivery {
  id: number;
  webhook_id: number;
  event: string;
  status: string;
  attempts: number;
  created_at: string;
  response_code?: number;
}

interface Integration {
  id: number;
  name: string;
  type: string;
  enabled: boolean;
  config: any;
  created_at: string;
}

const AVAILABLE_EVENTS = [
  'session.created',
  'session.started',
  'session.hibernated',
  'session.terminated',
  'session.failed',
  'user.created',
  'user.updated',
  'user.deleted',
  'dlp.violation',
  'recording.started',
  'recording.completed',
  'workflow.started',
  'workflow.completed',
  'collaboration.started',
  'compliance.violation',
  'security.alert',
  'scaling.event',
];

export default function Integrations() {
  const [currentTab, setCurrentTab] = useState(0);
  const [webhooks, setWebhooks] = useState<Webhook[]>([]);
  const [integrations, setIntegrations] = useState<Integration[]>([]);
  const [webhookDialog, setWebhookDialog] = useState(false);
  const [integrationDialog, setIntegrationDialog] = useState(false);
  const [deliveryDialog, setDeliveryDialog] = useState(false);
  const [selectedWebhook, setSelectedWebhook] = useState<Webhook | null>(null);
  const [deliveries, setDeliveries] = useState<WebhookDelivery[]>([]);

  const [webhookForm, setWebhookForm] = useState({
    name: '',
    url: '',
    secret: '',
    events: [] as string[],
    enabled: true,
  });

  const [integrationForm, setIntegrationForm] = useState({
    name: '',
    type: 'slack',
    config: {},
  });

  const handleCreateWebhook = () => {
    // TODO: API call to create webhook
    console.log('Create webhook:', webhookForm);
    setWebhookDialog(false);
  };

  const handleTestWebhook = (webhook: Webhook) => {
    // TODO: API call to test webhook
    console.log('Test webhook:', webhook.id);
  };

  const handleViewDeliveries = (webhook: Webhook) => {
    setSelectedWebhook(webhook);
    // TODO: Fetch deliveries for this webhook
    setDeliveryDialog(true);
  };

  const handleDeleteWebhook = (id: number) => {
    // TODO: API call to delete webhook
    setWebhooks(webhooks.filter((w) => w.id !== id));
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'success':
        return <SuccessIcon color="success" />;
      case 'failed':
        return <ErrorIcon color="error" />;
      case 'pending':
        return <PendingIcon color="warning" />;
      default:
        return <PendingIcon />;
    }
  };

  return (
    <Layout>
      <Box>
        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
          <Typography variant="h4" sx={{ fontWeight: 700 }}>
            Integration Hub
          </Typography>
          <Button
            variant="contained"
            startIcon={<AddIcon />}
            onClick={() => {
              setWebhookForm({ name: '', url: '', secret: '', events: [], enabled: true });
              setWebhookDialog(true);
            }}
          >
            New Webhook
          </Button>
        </Box>

        <Tabs value={currentTab} onChange={(_, v) => setCurrentTab(v)} sx={{ mb: 3 }}>
          <Tab label="Webhooks" />
          <Tab label="External Integrations" />
        </Tabs>

        {/* Webhooks Tab */}
        {currentTab === 0 && (
          <Card>
            <CardContent>
              <TableContainer>
                <Table>
                  <TableHead>
                    <TableRow>
                      <TableCell>Name</TableCell>
                      <TableCell>URL</TableCell>
                      <TableCell>Events</TableCell>
                      <TableCell>Status</TableCell>
                      <TableCell>Actions</TableCell>
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {webhooks.length === 0 ? (
                      <TableRow>
                        <TableCell colSpan={5} align="center">
                          <Typography color="text.secondary">No webhooks configured</Typography>
                        </TableCell>
                      </TableRow>
                    ) : (
                      webhooks.map((webhook) => (
                        <TableRow key={webhook.id}>
                          <TableCell>{webhook.name}</TableCell>
                          <TableCell>
                            <Typography variant="body2" sx={{ fontFamily: 'monospace', fontSize: '0.85rem' }}>
                              {webhook.url}
                            </Typography>
                          </TableCell>
                          <TableCell>
                            <Chip label={`${webhook.events.length} events`} size="small" />
                          </TableCell>
                          <TableCell>
                            <Chip
                              label={webhook.enabled ? 'Enabled' : 'Disabled'}
                              color={webhook.enabled ? 'success' : 'default'}
                              size="small"
                            />
                          </TableCell>
                          <TableCell>
                            <IconButton size="small" onClick={() => handleTestWebhook(webhook)} title="Test">
                              <TestIcon />
                            </IconButton>
                            <IconButton size="small" onClick={() => handleViewDeliveries(webhook)} title="History">
                              <HistoryIcon />
                            </IconButton>
                            <IconButton size="small" onClick={() => handleDeleteWebhook(webhook.id)} title="Delete">
                              <DeleteIcon />
                            </IconButton>
                          </TableCell>
                        </TableRow>
                      ))
                    )}
                  </TableBody>
                </Table>
              </TableContainer>
            </CardContent>
          </Card>
        )}

        {/* External Integrations Tab */}
        {currentTab === 1 && (
          <Card>
            <CardContent>
              <Grid container spacing={2}>
                {['Slack', 'Microsoft Teams', 'Discord', 'PagerDuty', 'Email'].map((type) => (
                  <Grid item xs={12} sm={6} md={4} key={type}>
                    <Card variant="outlined">
                      <CardContent>
                        <Typography variant="h6">{type}</Typography>
                        <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                          Connect {type} for notifications
                        </Typography>
                        <Button variant="outlined" size="small" fullWidth>
                          Configure
                        </Button>
                      </CardContent>
                    </Card>
                  </Grid>
                ))}
              </Grid>
            </CardContent>
          </Card>
        )}

        {/* Create/Edit Webhook Dialog */}
        <Dialog open={webhookDialog} onClose={() => setWebhookDialog(false)} maxWidth="md" fullWidth>
          <DialogTitle>Create Webhook</DialogTitle>
          <DialogContent>
            <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2, pt: 1 }}>
              <TextField
                label="Name"
                fullWidth
                value={webhookForm.name}
                onChange={(e) => setWebhookForm({ ...webhookForm, name: e.target.value })}
              />
              <TextField
                label="URL"
                fullWidth
                value={webhookForm.url}
                onChange={(e) => setWebhookForm({ ...webhookForm, url: e.target.value })}
                placeholder="https://example.com/webhook"
              />
              <TextField
                label="Secret (optional)"
                fullWidth
                type="password"
                value={webhookForm.secret}
                onChange={(e) => setWebhookForm({ ...webhookForm, secret: e.target.value })}
                helperText="Used for HMAC signature verification"
              />
              <FormControl fullWidth>
                <InputLabel>Events</InputLabel>
                <Select
                  multiple
                  value={webhookForm.events}
                  onChange={(e) =>
                    setWebhookForm({
                      ...webhookForm,
                      events: typeof e.target.value === 'string' ? [e.target.value] : e.target.value,
                    })
                  }
                  renderValue={(selected) => (
                    <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 0.5 }}>
                      {selected.map((value) => (
                        <Chip key={value} label={value} size="small" />
                      ))}
                    </Box>
                  )}
                >
                  {AVAILABLE_EVENTS.map((event) => (
                    <MenuItem key={event} value={event}>
                      {event}
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>
              <FormControlLabel
                control={
                  <Switch
                    checked={webhookForm.enabled}
                    onChange={(e) => setWebhookForm({ ...webhookForm, enabled: e.target.checked })}
                  />
                }
                label="Enabled"
              />
            </Box>
          </DialogContent>
          <DialogActions>
            <Button onClick={() => setWebhookDialog(false)}>Cancel</Button>
            <Button variant="contained" onClick={handleCreateWebhook}>
              Create
            </Button>
          </DialogActions>
        </Dialog>

        {/* Webhook Delivery History Dialog */}
        <Dialog open={deliveryDialog} onClose={() => setDeliveryDialog(false)} maxWidth="lg" fullWidth>
          <DialogTitle>Webhook Delivery History - {selectedWebhook?.name}</DialogTitle>
          <DialogContent>
            <TableContainer>
              <Table size="small">
                <TableHead>
                  <TableRow>
                    <TableCell>Status</TableCell>
                    <TableCell>Event</TableCell>
                    <TableCell>Attempts</TableCell>
                    <TableCell>Response</TableCell>
                    <TableCell>Time</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {deliveries.length === 0 ? (
                    <TableRow>
                      <TableCell colSpan={5} align="center">
                        <Typography color="text.secondary">No delivery history</Typography>
                      </TableCell>
                    </TableRow>
                  ) : (
                    deliveries.map((delivery) => (
                      <TableRow key={delivery.id}>
                        <TableCell>{getStatusIcon(delivery.status)}</TableCell>
                        <TableCell>{delivery.event}</TableCell>
                        <TableCell>{delivery.attempts}</TableCell>
                        <TableCell>{delivery.response_code || '-'}</TableCell>
                        <TableCell>{new Date(delivery.created_at).toLocaleString()}</TableCell>
                      </TableRow>
                    ))
                  )}
                </TableBody>
              </Table>
            </TableContainer>
          </DialogContent>
          <DialogActions>
            <Button onClick={() => setDeliveryDialog(false)}>Close</Button>
          </DialogActions>
        </Dialog>
      </Box>
    </Layout>
  );
}
