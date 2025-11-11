// BigBlueButton open source conferencing system - http://www.bigbluebutton.org/.
//
// Copyright (c) 2022 BigBlueButton Inc. and by respective authors (see below).
//
// This program is free software; you can redistribute it and/or modify it under the
// terms of the GNU Lesser General Public License as published by the Free Software
// Foundation; either version 3.0 of the License, or (at your option) any later
// version.
//
// Greenlight is distributed in the hope that it will be useful, but WITHOUT ANY
// WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A
// PARTICULAR PURPOSE. See the GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License along
// with Greenlight; if not, see <http://www.gnu.org/licenses/>.

import React from "react";
import { Card, Button } from "react-bootstrap";
import { useTranslation } from "react-i18next";
import {
  CurrencyDollarIcon,
  ChatBubbleLeftRightIcon,
  CheckCircleIcon,
  StarIcon,
} from "@heroicons/react/24/outline";

export default function SalesContactCard() {
  const { t } = useTranslation();

  const handleContactSales = () => {
    // You can replace this with your actual sales contact method
    // For example: window.open('mailto:sales@yourcompany.com', '_blank');
    // Or navigate to a contact form
    window.open(
      "mailto:ask@softylines.com?subject=Interest in Premium Features",
      "_blank"
    );
  };

  const features = [
    {
      text: "Unlimited meeting rooms",
      icon: <CheckCircleIcon className="hi-xs text-success me-2" />,
    },
    {
      text: "Advanced recording features",
      icon: <CheckCircleIcon className="hi-xs text-success me-2" />,
    },
    {
      text: "Priority support",
      icon: <CheckCircleIcon className="hi-xs text-success me-2" />,
    },
  ];

  return (
    <Card
      className="sales-card border-0 shadow-lg h-100 position-relative overflow-hidden"
      style={{
        background: `linear-gradient(135deg, #3c5de015 0%, #3c5de005 100%)`,
        border: `2px solid #3c5de020`,
      }}
    >
      {/* Decorative elements */}
      <div
        className="position-absolute top-0 end-0 p-3"
        style={{ opacity: 0.1 }}
      >
        <CurrencyDollarIcon className="hi-xl" style={{ color: "#3c5de0" }} />
      </div>

      <Card.Body className="p-4 position-relative">
        {/* Header */}
        <div className="text-center mb-4">
          <div
            className="rounded-circle d-inline-flex align-items-center justify-content-center mb-3"
            style={{
              width: "60px",
              height: "60px",
              backgroundColor: "#3c5de0",
              color: "white",
            }}
          >
            <StarIcon className="hi-l" />
          </div>
          <h4 className="fw-bold mb-2" style={{ color: "#3c5de0" }}>
            Upgrade to Premium
          </h4>
          <p className="text-muted mb-0">
            Get access to advanced features and priority support
          </p>
        </div>

        {/* Features list */}
        <div className="mb-4">
          <h6 className="fw-semibold mb-3">What you'll get:</h6>
          <div className="row g-2">
            {features.map((feature, index) => (
              <div key={index} className="col-12">
                <div className="d-flex align-items-center">
                  {feature.icon}
                  <span className="text-muted small">{feature.text}</span>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* CTA Section */}
        <div className="text-center">
          <Button
            onClick={handleContactSales}
            size="lg"
            className="w-100 mb-3 fw-semibold"
            style={{
              backgroundColor: "#3c5de0",
              borderColor: "#3c5de0",
              borderRadius: "12px",
              padding: "12px 24px",
              fontSize: "16px",
              boxShadow: `0 4px 12px #3c5de040`,
            }}
            onMouseEnter={(e) => {
              e.target.style.boxShadow = `0 6px 20px #3c5de050`;
            }}
            onMouseLeave={(e) => {
              e.target.style.boxShadow = `0 4px 12px #3c5de040`;
            }}
          >
            <ChatBubbleLeftRightIcon className="hi-s me-2" />
            Contact Sales
          </Button>

          <p className="small text-muted mb-0">
            Custom pricing available for teams and enterprises
          </p>
        </div>
      </Card.Body>

      {/* Bottom accent */}
      <div
        className="position-absolute bottom-0 start-0 w-100"
        style={{
          height: "4px",
          background: `linear-gradient(90deg, #3c5de0, #3c5de080)`,
        }}
      />
    </Card>
  );
}
